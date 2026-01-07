# OIP Backend - 架构说明

## 一、整体架构

```
┌─────────────────────────────────────────────────┐
│              用户/商家系统                        │
└──────────┬──────────────────────────┬───────────┘
           │ HTTP API                 │
           ↓                          │
    ┌──────────────┐                 │
    │   dpmain     │                 │
    │  (API 服务)   │                 │
    └──────┬───────┘                 │
           │                          │
           │ 1. 订阅 Redis            │ 4. 轮询查询
           │ 2. 推送 Lmstfy           │
           │ 3. Smart Wait (10s)      │
           ↓                          ↓
    ┌──────────────────────────────────────┐
    │  Redis Pub/Sub + MySQL              │
    │  channel: diagnosis:result:{id}     │
    └──────────────┬──────────────────────┘
                   ↑
                   │ 5. 发布结果
    ┌──────────────┴──────────┐
    │      dpsync             │
    │   (Worker 服务)          │
    │                         │
    │  CompositeHandler       │
    │  ├─ ShippingCalculator  │
    │  └─ AnomalyChecker      │
    └─────────────────────────┘
           ↑
           │ 消费 Lmstfy
    ┌──────┴─────────┐
    │  oip_order_    │
    │  diagnose      │
    │  (消息队列)     │
    └────────────────┘
```

## 二、模块职责

### common（共享内核）

**职责**：提供共享的数据模型和 DAO 层

**包含**：
- `entity/`: GORM 数据模型
  - `order.go`: Order 实体（包含 diagnose_result）
  - `account.go`: Account 实体
- `model/`: 诊断结果结构体
  - `diagnosis_result.go`: DiagnosisResultData, DiagnosisItem
  - `shipping_result.go`: ShippingResult, ShippingRate
  - `anomaly_result.go`: AnomalyResult, AnomalyItem
  - `response.go`: Response, MetaInfo, ErrorDetail
- `dao/`: 数据访问层
  - `order_dao.go`: OrderDAO（Create, GetByID, UpdateDiagnoseResult）
  - `account_dao.go`: AccountDAO（Create, GetByID）

**依赖**：
- gorm.io/gorm
- gorm.io/datatypes

### dpmain（同步 API 服务）

**职责**：提供 HTTP API，实现 Smart Wait 机制

**包含**：
- `cmd/apiserver/`: 程序入口
  - `main.go`: 启动 Gin 服务器，注册路由
- `internal/api/`: HTTP Handlers
  - `order_handler.go`: CreateOrder, GetOrder
  - `account_handler.go`: CreateAccount, GetAccount
- `internal/service/`: 业务服务层
  - `order_service.go`: OrderService（业务编排）
- `internal/middleware/`: 中间件
  - `cors.go`: CORS 跨域处理
- `pkg/config/`: 配置管理
  - `config.go`: Config 结构体和 Load 方法
- `pkg/redis/`: Redis 客户端封装
  - `client.go`: Subscribe, Publish

**依赖**：
- github.com/gin-gonic/gin（HTTP 框架）
- github.com/redis/go-redis/v9（Redis 客户端）
- oip/common（共享内核）

**核心流程**：
1. 接收订单请求（POST /api/v1/orders?wait=10）
2. 验证 account_id 存在
3. 地址格式校验（正则）
4. 落库（status=DIAGNOSING）
5. 订阅 Redis channel: `diagnosis:result:{order_id}`
6. 推送 Lmstfy 队列
7. Smart Wait（10s 超时）
8. 返回结果：200（诊断完成）或 3001（诊断中）

### dpsync（异步 Worker 服务）

**职责**：消费 Lmstfy 队列，执行诊断任务

**包含**：
- `cmd/worker/`: 程序入口
  - `main.go`: 启动 Worker，监听退出信号
- `internal/worker/`: Worker 核心逻辑
  - `worker.go`: Worker（Start, Stop）
- `internal/handlers/`: 业务处理器
  - `composite_handler.go`: CompositeHandler（Handle）
  - `shipping_calculator.go`: ShippingCalculator（Calculate）
  - `anomaly_checker.go`: AnomalyChecker（Check）
- `pkg/config/`: 配置管理
  - `config.go`: Config 结构体和 Load 方法
- `pkg/lmstfy/`: Lmstfy 客户端封装
  - `client.go`: Consume, Publish
- `config/`: 配置文件
  - `worker.yaml`: Worker 配置示例

**依赖**：
- github.com/bitleak/lmstfy（消息队列客户端）
- github.com/redis/go-redis/v9（Redis 客户端）
- oip/common（共享内核）

**核心流程**：
1. 消费 Lmstfy 队列（oip_order_diagnose）
2. 解析 job data（order_id, account_id）
3. 查询订单数据
4. 执行 CompositeHandler：
   - 异常检测（AnomalyChecker）
   - 费率计算（ShippingCalculator）
5. 组装 DiagnosisResultData
6. 更新 DB（status=DIAGNOSED, diagnose_result=JSON）
7. 发布 Redis 通知（channel: diagnosis:result:{order_id}）

## 三、数据流转

### 创建订单（同步返回）

```
用户 → POST /api/v1/orders?wait=10
  ↓
dpmain:
  1. 验证 account_id 存在（SELECT FROM accounts）
  2. 地址格式校验（正则）
  3. 落库（INSERT INTO orders, status=DIAGNOSING）
  4. 订阅 Redis: diagnosis:result:{order_id}
  5. 推送 Lmstfy: {"order_id": "...", "account_id": 1}
  6. Smart Wait（10s）:
     - 收到 Redis 消息 → 返回 200 + 诊断结果
     - 超时 → 返回 3001 + poll_url
```

### 异步诊断（dpsync）

```
dpsync:
  1. 消费 Lmstfy 队列
  2. 查询订单（SELECT FROM orders WHERE id=?）
  3. CompositeHandler.Handle():
     - AnomalyChecker.Check() → AnomalyResult
     - ShippingCalculator.Calculate() → ShippingResult
     - 组装 DiagnosisResultData
  4. 更新 DB:
     UPDATE orders SET
       status='DIAGNOSED',
       diagnose_result='{"items": [...]}'
     WHERE id=?
  5. 发布 Redis:
     PUBLISH diagnosis:result:{order_id} {"order_id": "...", "data": {...}}
```

### 轮询查询（超时场景）

```
用户 → GET /api/v1/orders/{id}
  ↓
dpmain:
  1. 查询订单（SELECT FROM orders WHERE id=?）
  2. 返回完整数据（包含 diagnose_result）
```

## 四、关键设计决策

### 1. 单表设计（简化）
- 不使用独立的 `diagnoses` 表
- 诊断结果直接存储在 `orders.diagnose_result`（JSON）
- 减少 JOIN 查询，提升性能

### 2. Redis 独立 channel
- 每个订单独立 channel: `diagnosis:result:{order_id}`
- 避免多实例串消息
- 请求结束后自动取消订阅

### 3. 数组 + type 扩展
```json
{
  "items": [
    {"type": "shipping", "status": "SUCCESS", "data_json": {...}},
    {"type": "anomaly", "status": "SUCCESS", "data_json": {...}}
  ]
}
```
- 扩展新诊断类型时只需追加数组元素
- 不需要修改表结构

### 4. 框架 vs 业务分离
- 框架负责：消息消费、状态更新、通知
- 业务负责：具体诊断逻辑
- `CompositeHandler.Handle()` 返回 `(resultData, error)`

### 5. 预留扩展点
- 所有业务逻辑标记为 `TODO`
- 便于审核架构后再实现具体逻辑

## 五、环境依赖

### 开发环境
- Go 1.21+
- Make

### 运行时依赖
- MySQL 8.0（订单数据存储）
- Redis 7.0（Pub/Sub 通知）
- Lmstfy（消息队列）

### 启动基础设施
```bash
docker-compose up -d
```

## 六、验证步骤

### 1. 环境检查
```bash
go version  # 确保 1.21+
go env GOPATH  # 确保指向有权限的目录
```

### 2. 构建验证
```bash
cd /Users/cooperswang/GolandProjects/awesomeProject/oip_backend
./scripts/verify.sh
```

### 3. 预期输出
```
✓ common 模块验证成功
✓ dpmain 模块构建成功: dpmain/bin/dpmain-apiserver
✓ dpsync 模块构建成功: dpsync/bin/dpsync-worker
✓ Workspace 同步成功
```

### 4. 启动服务
```bash
# 终端 1：启动 API 服务
make run-dpmain

# 终端 2：启动 Worker 服务
make run-dpsync
```

## 七、下一步开发任务

### Week 1: 基础设施
- [ ] 实现 `common/dao` 的 CRUD 逻辑
- [ ] 实现 `dpmain/pkg/redis` 的 Pub/Sub 封装
- [ ] 实现 `dpsync/pkg/lmstfy` 的消息消费

### Week 2: 核心业务
- [ ] 实现 `dpmain/internal/api` 的 HTTP 接口
- [ ] 实现 Smart Wait 机制
- [ ] 实现 `dpsync/internal/handlers` 的诊断逻辑

### Week 3: Mock 数据
- [ ] 实现 ShippingCalculator（Mock 费率）
- [ ] 实现 AnomalyChecker（规则引擎）

### Week 4: 测试与优化
- [ ] 端到端测试
- [ ] 压力测试
- [ ] 性能优化
