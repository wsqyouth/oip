# OIP 架构核心原则

## 系统概览

### 服务架构
- **dpmain**: 订单主服务（同步 API 服务，端口 7777）
  - HTTP API Server（Gin）
  - Callback Consumer（异步消费诊断回调）
- **dpsync**: 数据同步服务（异步 Worker 服务，端口 7778）
  - Worker（消费诊断任务队列）
  - CompositeHandler（执行诊断逻辑）
- **common**: 共享内核（Shared Kernel）
  - Entity（GORM 数据模型）
  - Model（诊断结果结构体）
  - DAO（数据访问层）

### 异步框架
- **Smart Wait 机制**: dpmain Hold 连接 10s 等待诊断结果
- **消息队列**: Lmstfy
  - `oip_order_diagnose`: 诊断任务队列
  - `oip_order_diagnose_callback`: 诊断回调队列
- **通知机制**: Redis Pub/Sub
  - Channel 格式: `diagnosis:result:{order_id}`

## 关键设计决策

### 1. 为什么选择双服务架构？

**职责分离**：
- **dpmain**: 处理 HTTP 请求，提供低延迟响应
- **dpsync**: 处理计算密集型任务，避免阻塞 HTTP 请求

**扩展性**：
- 可独立扩展 dpmain（应对流量高峰）
- 可独立扩展 dpsync（应对诊断任务积压）

### 2. 为什么使用 Smart Wait？

**用户体验**：
- 大部分诊断任务可在 10s 内完成
- 直接返回结果，避免轮询

**降级策略**：
- 超时返回 3001 Processing
- 前端可轮询或 WebSocket 通知

### 3. 为什么使用单表设计？

**简化查询**：
- 诊断结果直接存储在 `orders.diagnose_result` (JSON)
- 避免 JOIN 查询，提升性能

**扩展性**：
- JSON 结构灵活，新增诊断类型无需修改表结构
- 使用 `items` 数组 + `type` 字段支持多种诊断类型

### 4. 异步任务如何流转？

```
商家系统
  ↓ POST /api/v1/orders?wait=10
dpmain
  ├─ 订单落库（status=DIAGNOSING）
  ├─ 订阅 Redis: diagnosis:result:{order_id}
  ├─ 推送 Lmstfy: oip_order_diagnose
  └─ Smart Wait (10s)
      ↓
      10s 内收到 Redis 消息 → 200 OK
      超时 → 3001 Processing
  ↓
dpsync
  ├─ 消费 Lmstfy: oip_order_diagnose
  ├─ 执行 CompositeHandler
  │   ├─ ShippingCalculator（费率计算）
  │   └─ AnomalyChecker（异常检测）
  ├─ 更新数据库（status=DIAGNOSED）
  └─ 发布 Redis: diagnosis:result:{order_id}
```

## 代码组织规范

### DDD 分层架构（dpmain）

```
domains/               # 领域层（业务逻辑）
  ├── entity/         # 聚合根和值对象
  ├── apimodel/       # API 模型（DTO）
  ├── modules/        # 业务编排（组合 services）
  ├── repo/           # 仓储接口（只定义）
  └── services/       # 领域服务（复杂逻辑）

infra/                # 基础设施层（技术实现）
  ├── persistence/    # 持久化实现
  └── mq/             # 消息队列

server/               # 服务器层（HTTP 适配）
  ├── handlers/       # HTTP 处理器
  ├── routers/        # 路由配置
  └── middlewares/    # 中间件
```

### 依赖方向

**核心原则**：依赖指向领域层，领域层不依赖外部

```
server/handlers
  ↓ 调用
modules/mdorder
  ↓ 调用
services/svorder + repo/rporder (接口)
  ↑ 实现
infra/persistence/mysql/order_repo_impl
```

### 命名约定

| 前缀 | 含义 | 示例 | 说明 |
|------|------|------|------|
| `et` | Entity | `etorder.Order` | 聚合根和值对象 |
| `md` | Module | `mdorder.OrderModule` | 业务编排层 |
| `rp` | Repository | `rporder.OrderRepository` | 仓储接口 |
| `sv` | Service | `svorder.OrderService` | 领域服务 |

## 数据流转

### HTTP 请求流（同步）

```
HTTP Request
  ↓ 解析参数
handlers/order/create.go
  ↓ 调用模块
modules/mdorder/order_module.go
  ↓ 调用服务
services/svorder/order_service.go
  ↓ 调用仓储接口
repo/rporder/order_repository.go (接口)
  ↓ 实现
infra/persistence/mysql/order_repo_impl.go
  ↓ 调用 DAO
common/dao/order_dao.go
  ↓ 操作数据库
MySQL
```

### 消息消费流（异步）

```
Lmstfy Queue: oip_order_diagnose
  ↓ 消费
dpsync/worker/worker.go
  ↓ 调用处理器
handlers/composite_handler.go
  ├─ shipping_calculator.go（费率计算）
  └─ anomaly_checker.go（异常检测）
  ↓ 组装结果
DiagnosisResultData
  ↓ 更新数据库
common/dao/order_dao.go
  ↓ 发布通知
Redis Pub/Sub: diagnosis:result:{order_id}
```

## 性能优化原则

### 1. 数据库优化
- **索引**:
  - `uk_account_merchant` (account_id, merchant_order_no)：幂等性检查
  - `idx_account_status` (account_id, status)：订单列表查询
  - `idx_created_at`：按时间排序
- **单表设计**: 避免 JOIN 查询
- **JSON 字段**: 灵活存储诊断结果，减少表关联

### 2. Redis 优化
- **独立 channel**: 每个订单独立 channel，避免多实例串消息
- **自动过期**: 订阅结束后自动取消

### 3. 异步优化
- **批量处理**: dpsync 可配置并发数（默认 5）
- **重试机制**: Lmstfy 支持任务重试（默认 3 次）

## 安全性原则

### 1. 输入验证
- **地址校验**: 正则表达式验证地址格式
- **账号验证**: 验证 account_id 存在
- **幂等性**: 基于 merchant_order_no 去重

### 2. 错误处理
- **堆栈追踪**: 使用 `errors.Wrap` 保留堆栈
- **错误分类**: 业务错误 vs 系统错误
- **统一响应**: MetaInfo 统一错误码

### 3. 配置管理
- **禁止硬编码**: 所有配置从 config.yaml 读取
- **敏感信息**: 数据库密码等使用环境变量

## 扩展性设计

### 1. 诊断类型扩展

**当前支持**：
- `shipping`: 物流费率计算
- `anomaly`: 异常检测

**扩展方式**：
```json
{
  "items": [
    {"type": "shipping", "status": "SUCCESS", "data_json": {...}},
    {"type": "anomaly", "status": "SUCCESS", "data_json": {...}},
    {"type": "risk", "status": "SUCCESS", "data_json": {...}}  // 新增
  ]
}
```

只需：
1. 在 `common/model/` 下定义新的结果结构体
2. 在 `dpsync/handlers/composite_handler.go` 中添加新的处理器
3. 无需修改表结构

### 2. 新增服务
- 新增 `dpnotify` 服务处理通知推送
- 新增 `dpreport` 服务处理数据统计

### 3. 水平扩展
- dpmain: 无状态，可任意扩展
- dpsync: 通过 Lmstfy 队列分发任务，可任意扩展

## 常见误区

### ❌ 误区 1: 在 handler 层直接操作数据库
```go
// 错误
func CreateOrder(c *gin.Context) {
    db.Create(&order)
}
```

✅ 正确：调用模块
```go
func CreateOrder(c *gin.Context) {
    orderModule.CreateOrder(ctx, payload)
}
```

### ❌ 误区 2: module 调用 module
```go
// 错误
type OrderModule struct {
    accountModule mdaccount.Module // module 不应该调用 module
}
```

✅ 正确：service 编排多个 module
```go
type OrderService struct {
    orderModule   mdorder.Module
    accountModule mdaccount.Module  // service 编排 module
}
```

### ❌ 误区 3: 硬编码配置
```go
// 错误
port := 7777
```

✅ 正确：从配置读取
```go
port := cfg.Server.Port
```

### ❌ 误区 4: 忘记事务回滚
```go
// 错误
tx := db.Begin()
if err := dao.Create(tx, order); err != nil {
    return err // 忘记回滚
}
tx.Commit()
```

✅ 正确：defer 回滚
```go
tx := db.Begin()
defer func() {
    if err != nil {
        tx.Rollback()
    }
}()
```
