# OIP Backend - 生产级架构重构总结

**重构时间**: 2025-12-23
**状态**: ✅ 架构框架已完成，待注入依赖和实现业务逻辑

---

## 一、架构对比

### **重构前（简化版）**
```
dpmain/
├── internal/
│   ├── api/          # 扁平的 HTTP Handlers
│   ├── service/      # 扁平的业务服务
│   └── middleware/
└── pkg/
```
**问题**：
- 缺少 DDD 分层
- 无法清晰区分领域对象和 DTO
- 基础设施未统一管理

### **重构后（生产级）**
```
dpmain/
├── internal/app/
│   ├── domains/           # 【领域层】DDD 核心
│   │   ├── entity/        # 实体（etorder, etaccount）
│   │   ├── repo/          # 仓储接口（rporder, rpaccount）
│   │   ├── services/      # 领域服务（svorder, svdiagnosis）
│   │   ├── modules/       # 业务编排（mdorder, mdaccount）
│   │   └── apimodel/      # DTO（request, response）
│   ├── infra/             # 【基础设施层】
│   │   ├── persistence/   # 仓储实现（mysql, redis）
│   │   └── mq/            # 消息队列（lmstfy）
│   ├── server/            # 【服务器层】
│   │   ├── handlers/      # HTTP 处理器（按模块分组）
│   │   ├── routers/       # 路由配置
│   │   └── middlewares/   # 中间件
│   └── pkg/               # 通用包（errorx, ginx, logger）
```

---

## 二、关键架构决策

### 1. DDD 分层架构
- **Entity（实体层）**: 纯领域对象，封装业务规则
- **Repository（仓储层）**: 接口定义与实现分离
- **Service（服务层）**: 复杂业务逻辑
- **Module（模块层）**: 业务编排，组合多个服务

### 2. 命名规范
| 前缀 | 含义 | 示例 |
|------|------|------|
| `et` | Entity（实体） | `etorder.Order` |
| `md` | Module（模块） | `mdorder.OrderModule` |
| `rp` | Repository（仓储） | `rporder.OrderRepository` |
| `sv` | Service（服务） | `svorder.OrderService` |

### 3. 领域对象 vs GORM 模型
```
dpmain/domains/entity/etorder/order.go  (领域对象 - 纯业务逻辑)
            ↓ 转换
common/entity/order.go                  (GORM 模型 - 数据库映射)
```
- **分离关注点**: 领域层不依赖任何框架（GORM, Gin）
- **转换层**: 在 `infra/persistence` 进行对象转换

### 4. DTO 与 Entity 分离
- **apimodel/request**: HTTP 请求 DTO
- **apimodel/response**: HTTP 响应 DTO
- **entity**: 领域对象（不直接暴露给外部）

### 5. 基础设施统一管理
- **infra/persistence**: 所有数据库操作
- **infra/mq**: 所有消息队列操作
- **便于替换**: MySQL → PostgreSQL, Lmstfy → Kafka

---

## 三、目录结构详解

### **domains/（领域层）**
```
domains/
├── entity/                    # 实体（纯领域对象）
│   ├── etorder/
│   │   └── order.go           # Order 聚合根（工厂方法、领域行为）
│   ├── etaccount/
│   │   └── account.go         # Account 实体
│   └── etprimitive/
│       └── types.go           # 基础类型和值对象
│
├── apimodel/                  # API 模型（DTO）
│   ├── request/
│   │   ├── create_order_request.go
│   │   └── create_account_request.go
│   └── response/
│       ├── order_response.go
│       └── account_response.go
│
├── modules/                   # 领域模块（业务编排）
│   ├── mdorder/
│   │   └── order_module.go    # CreateOrder, GetOrder（编排逻辑）
│   └── mdaccount/
│       └── account_module.go
│
├── repo/                      # 仓储接口（只定义）
│   ├── rporder/
│   │   └── order_repo.go      # OrderRepository interface
│   └── rpaccount/
│       └── account_repo.go    # AccountRepository interface
│
└── services/                  # 领域服务（复杂业务逻辑）
    ├── svorder/
    │   └── order_service.go   # ValidateAddressFormat, CheckDuplicate
    └── svdiagnosis/
        └── diagnosis_service.go # SubscribeResult, PublishToDiagnoseQueue
```

### **infra/（基础设施层）**
```
infra/
├── persistence/               # 持久化实现
│   ├── mysql/
│   │   ├── order_repo_impl.go     # 实现 rporder.OrderRepository
│   │   └── account_repo_impl.go   # 实现 rpaccount.AccountRepository
│   └── redis/
│       └── pubsub_client.go       # Redis Pub/Sub 封装
│
└── mq/                        # 消息队列
    └── lmstfy/
        └── client.go              # Lmstfy 客户端封装
```

### **server/（服务器层）**
```
server/
├── handlers/                  # HTTP 处理器（按模块分组）
│   ├── order/
│   │   ├── handler.go         # OrderHandler 结构体
│   │   ├── create.go          # POST /api/v1/orders
│   │   └── get.go             # GET /api/v1/orders/:id
│   └── account/
│       ├── handler.go
│       ├── create.go          # POST /api/v1/accounts
│       └── get.go             # GET /api/v1/accounts/:id
│
├── routers/
│   └── router.go              # SetupRoutes（统一路由注册）
│
└── middlewares/
    ├── cors.go                # CORS 跨域
    ├── logger.go              # 日志中间件
    └── error.go               # 统一错误处理
```

### **pkg/（通用包）**
```
pkg/
├── errorx/
│   └── errors.go              # 业务错误定义
├── ginx/
│   └── response.go            # 统一响应格式（Success, Error, Processing）
└── logger/
    └── logger.go              # 日志接口（可扩展为 zap）
```

---

## 四、数据流转示例

### **创建订单流程**
```
1. HTTP Request
   POST /api/v1/orders?wait=10

2. server/handlers/order/create.go
   - 解析 request.CreateOrderRequest（DTO）
   - 调用 orderModule.CreateOrder()

3. domains/modules/mdorder/order_module.go
   - 验证 account_id 存在（调用 accountRepo）
   - 检查订单重复（调用 orderService）
   - 验证地址格式（调用 orderService）
   - 创建 etorder.Order 领域对象
   - 调用 orderRepo.Create（仓储接口）
   - 发布到诊断队列（调用 diagnosisService）
   - 订阅 Redis 结果（Smart Wait）

4. infra/persistence/mysql/order_repo_impl.go
   - 将 etorder.Order 转换为 common/entity.Order（GORM 模型）
   - 执行 db.Create()

5. infra/mq/lmstfy/client.go
   - 发布到 oip_order_diagnose 队列

6. infra/persistence/redis/pubsub_client.go
   - 订阅 diagnosis:result:{order_id}
   - 等待 10s 超时

7. Response
   - 收到结果: 200 OK + 诊断数据
   - 超时: 3001 Processing + poll_url
```

---

## 五、与 dpsync 的协作

```
dpmain (同步服务):
  - domains/entity/etorder/       → 定义 Order 领域对象
  - domains/repo/rporder/         → 定义 OrderRepository 接口
  - infra/persistence/mysql/      → 实现 OrderRepository

dpsync (异步服务):
  - 引用 common/entity/          → 使用 GORM 模型操作 DB
  - 引用 common/model/           → 使用共享的诊断结果结构体
  - 内部架构：经典的订阅消费模式
```

---

## 六、文件统计

### **dpmain 模块**
- **Go 文件**: 40+ 个
- **目录层级**: 4 层（domains, infra, server, pkg）
- **代码行数**: ~1500 行（框架代码，业务逻辑待实现）

### **关键文件清单**
```
dpmain/
├── cmd/apiserver/main.go                               # 程序入口
├── internal/app/
│   ├── domains/
│   │   ├── entity/etorder/order.go                     # Order 聚合根
│   │   ├── repo/rporder/order_repo.go                  # 仓储接口
│   │   ├── services/svorder/order_service.go           # 领域服务
│   │   ├── modules/mdorder/order_module.go             # 业务编排
│   │   └── apimodel/request/create_order_request.go    # 请求 DTO
│   ├── infra/
│   │   ├── persistence/mysql/order_repo_impl.go        # 仓储实现
│   │   └── persistence/redis/pubsub_client.go          # Redis 客户端
│   └── server/
│       ├── handlers/order/create.go                    # HTTP 处理器
│       └── routers/router.go                           # 路由配置
├── go.mod                                              # 模块依赖
├── Makefile                                            # 构建任务
└── README.md                                           # 架构说明
```

---

## 七、快速验证

### **构建测试**
```bash
cd /Users/cooperswang/GolandProjects/awesomeProject/oip_backend/dpmain

# 查看架构
make arch

# 构建
make build

# 运行
make run

# 测试 API
curl http://localhost:8080/health
curl http://localhost:8080/architecture
```

### **预期输出**
```json
{
  "status": "ok",
  "service": "dpmain",
  "message": "架构框架已就绪，待注入依赖"
}
```

---

## 八、下一步开发任务

### **Phase 1: 依赖注入（Week 1）**
- [ ] 引入 Wire（Google 依赖注入工具）
- [ ] 实现 `cmd/apiserver/wire.go`
- [ ] 注入所有依赖（repo, service, module, handler）

### **Phase 2: 基础设施实现（Week 1-2）**
- [ ] 实现 `infra/persistence/mysql` 的完整转换逻辑
- [ ] 实现 `infra/persistence/redis` 的 Smart Wait
- [ ] 实现 `infra/mq/lmstfy` 的发布逻辑

### **Phase 3: 业务逻辑实现（Week 2-3）**
- [ ] 实现 `domains/services` 的校验逻辑
- [ ] 实现 `domains/modules` 的编排逻辑
- [ ] 实现 `server/handlers` 的 HTTP 逻辑

### **Phase 4: 测试与优化（Week 3-4）**
- [ ] 单元测试（覆盖率 > 80%）
- [ ] 端到端测试
- [ ] 性能优化

---

## 九、架构优势

### **1. 清晰的职责划分**
- 领域层：纯业务逻辑，不依赖任何框架
- 基础设施层：技术实现细节，易于替换
- 服务器层：HTTP 适配，与领域解耦

### **2. 高可测试性**
- 接口与实现分离（repo 接口 + infra 实现）
- 可使用 Mock 对象进行单元测试
- 领域对象无外部依赖，易于测试

### **3. 高可扩展性**
- 新增诊断类型：只需添加新的 service 和 module
- 替换数据库：只需修改 infra/persistence 实现
- 替换消息队列：只需修改 infra/mq 实现

### **4. 符合生产级标准**
- 参考现有生产代码的目录结构
- 清晰的命名规范（et/md/rp/sv 前缀）
- 完整的 DDD 分层

---

## 十、总结

✅ **已完成**：
1. 完整的 DDD 分层架构
2. 40+ 个框架文件（所有业务逻辑标记为 TODO）
3. 清晰的命名规范和目录结构
4. 领域对象与 GORM 模型分离
5. DTO 与 Entity 分离
6. 基础设施层统一管理

⏳ **待完成**：
1. 依赖注入（Wire）
2. 具体业务逻辑实现
3. 单元测试和集成测试

🎯 **目标达成**：
- ✅ dpmain 使用生产级 DDD 架构
- ✅ 框架搭建完成，待注入依赖
- ✅ 可独立构建和运行（健康检查通过）
