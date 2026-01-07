# DPSYNC 架构设计文档

## 一、分层架构概览

DPSYNC 采用清晰的分层架构，从下到上分为：**框架层 → Worker 层 → 应用层 → 业务层 → 基础设施层**

```
┌─────────────────────────────────────────────────────────────┐
│                     应用入口（cmd/）                          │
│                    main.go - 启动 Manager                     │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                   Worker 层（internal/worker/）               │
│  • Manager - 多 Worker 管理 + 依赖注入                        │
│  • WorkerInstance - 封装 Subscriber + Processor               │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                  框架层（internal/framework/）                │
│  • Subscriber - 主动拉取消息（容错重试、速率控制）              │
│  • Processor - 被动处理消息（Drain 模式）                      │
│  • MessageSource 接口 - 消息队列抽象                          │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                 应用层（internal/domains/）                   │
│  • processor.go - GetProcess 统一入口                         │
│  • handler_map.go - HandlerMap 路由表                         │
│  • common/ - Job、Response、HandlerServ 抽象                  │
│  • handlers/ - 业务 Handler（如 DiagnoseHandler）             │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│                 业务层（internal/business/）                  │
│  • CompositeHandler - 复合诊断处理器                          │
│  • ShippingCalculator - 物流费率计算                          │
│  • AnomalyChecker - 异常检测                                  │
│  • DiagnosisService - 诊断服务（协调业务+DB+Redis）            │
└─────────────────────────────────────────────────────────────┘
                              ↓
┌─────────────────────────────────────────────────────────────┐
│               基础设施层（pkg/infra/）                         │
│  • mysql/order_dao.go - 订单数据访问                          │
│  • redis/pubsub.go - Redis 发布订阅                           │
└─────────────────────────────────────────────────────────────┘
```

---

## 二、目录结构与职责

### 1. 框架层：`internal/framework/`

**职责**：提供可复用的消息消费框架，与业务逻辑完全解耦

```
internal/framework/
├── subscriber.go       # Subscriber - 主动拉取消息
├── processor.go        # Processor - 被动处理消息（Drain 模式）
├── interfaces.go       # MessageSource 接口定义
├── types.go            # Message、ProcessResult 类型
└── config.go           # SubscriberConfig、ProcessorConfig
```

**核心特性**：
- ✅ Subscriber/Processor 分离
- ✅ 4 步优雅退出 + Drain 模式（零消息丢失）
- ✅ 容错重试 + 速率控制
- ✅ Deadlock 防护（select + ctx.Done()）

**依赖**：仅依赖 `pkg/lmstfyx`（接口定义）和 `pkg/logger`

---

### 2. Worker 层：`internal/worker/`

**职责**：Worker 生命周期管理和依赖注入

```
internal/worker/
├── worker.go           # WorkerInstance - 封装 Subscriber + Processor
└── manager.go          # Manager - 多 Worker 管理 + 依赖注入
```

**核心职责**：
- ✅ 创建并管理 Worker 实例
- ✅ 初始化依赖（OrderDAO、Redis PubSub、DiagnosisService）
- ✅ 通过 GetProcess 将依赖注入到 Context
- ✅ 优雅关闭所有 Worker 和连接

**依赖**：
- 向下依赖：`framework/`、`pkg/lmstfy/`
- 向上依赖：`domains/`、`business/`、`pkg/infra/`

---

### 3. 应用层：`internal/domains/`

**职责**：业务路由、Handler 管理、请求响应抽象

```
internal/domains/
├── processor.go                # GetProcess - 统一入口（路由分发）
├── handler_map.go              # HandlerMap - ActionType → Handler 映射
│
├── common/                     # 通用组件
│   ├── job/                    # Job 标准结构
│   │   └── job_entity.go
│   ├── response/               # Response 抽象
│   │   ├── response.go
│   │   └── diagnosis_result.go
│   └── handler_serv.go         # HandlerServ 接口
│
└── handlers/                   # 业务 Handler 实现
    └── order/
        └── diagnose/
            ├── handler.go      # DiagnoseHandler
            └── testcase/
                └── diagnose.json
```

**核心职责**：
- ✅ GetProcess 统一入口：解析 Job → 路由 → 调用 Handler → 错误处理
- ✅ HandlerMap 路由表：静态映射 ActionType → Handler
- ✅ Response + ResultI 抽象：统一响应结构
- ✅ DiagnoseHandler：订单诊断的应用逻辑

**设计原则**：
- **应用层只负责路由和协调**，不包含核心业务逻辑
- Handler 调用业务层（`business/`）完成实际业务处理
- 通过 Context 传递依赖（DiagnosisService）

---

### 4. 业务层：`internal/business/`

**职责**：核心业务逻辑实现，可独立测试

```
internal/business/
├── composite_handler.go       # 复合诊断处理器（组装结果）
├── shipping_calculator.go     # 物流费率计算器（Mock）
├── anomaly_checker.go         # 异常检测器（规则引擎）
└── diagnosis_service.go       # 诊断服务（协调业务+DB+Redis）
```

**核心职责**：
- ✅ **CompositeHandler**：组装 ShippingCalculator 和 AnomalyChecker 的结果
- ✅ **ShippingCalculator**：计算物流费率（确定性 Mock）
- ✅ **AnomalyChecker**：异常检测规则引擎
- ✅ **DiagnosisService**：协调诊断流程 + 数据持久化 + 事件通知

**设计原则**：
- **纯业务逻辑，不依赖框架**
- 可以在单元测试中直接调用
- 依赖注入（通过构造函数传入 OrderDAO、Redis PubSub）

---

### 5. 基础设施层：`pkg/infra/`

**职责**：提供数据访问和外部服务集成

```
pkg/infra/
├── mysql/
│   └── order_dao.go           # OrderDAO - 订单数据访问
└── redis/
    └── pubsub.go              # Redis Pub/Sub - 消息发布
```

**核心职责**：
- ✅ **OrderDAO**：更新订单诊断结果到 MySQL
- ✅ **Redis PubSub**：发布诊断完成通知到 `order_diagnosis_complete` 频道

**设计原则**：
- 封装底层技术细节（GORM、Redis SDK）
- 对外提供简洁的业务接口
- 可替换实现（便于测试和迁移）

---

## 三、核心数据流

```
lmstfy 队列消息
    ↓
Subscriber.loop() 拉取消息（多并发）
    ↓
发送到 inputChan（缓冲区）
    ↓
Processor.loop() 接收消息（多并发）
    ↓
调用 lmstfyx.Proc（即 GetProcess）
    ↓
┌────────────────────────────────────────────────────┐
│ GetProcess(ctx, job, diagnosisService)             │
│  1. parseJob() - 解析 Job 结构                     │
│  2. 注入 trace_id、diagnosisService 到 Context     │
│  3. HandlerMap 路由到 DiagnoseHandler              │
└────────────────────────────────────────────────────┘
    ↓
┌────────────────────────────────────────────────────┐
│ DiagnoseHandler.GetProcess()                       │
│  1. 解析 payload（order_id, account_id）           │
│  2. 从 Context 获取 DiagnosisService               │
│  3. 调用 DiagnosisService.ExecuteDiagnosis()       │
└────────────────────────────────────────────────────┘
    ↓
┌────────────────────────────────────────────────────┐
│ DiagnosisService.ExecuteDiagnosis()                │
│  ├─ CompositeHandler.Diagnose()                    │
│  │   ├─ ShippingCalculator.Calculate()             │
│  │   │   └─ 返回 ShippingResult（费率 + 推荐）     │
│  │   └─ AnomalyChecker.Check()                     │
│  │       └─ 返回 AnomalyResult（异常列表）         │
│  │                                                  │
│  ├─ OrderDAO.UpdateDiagnosisResult()               │
│  │   └─ 更新 orders 表（status + diagnose_result）│
│  │                                                  │
│  └─ RedisPubSub.PublishDiagnosisComplete()         │
│      └─ 发布到 order_diagnosis_complete 频道       │
└────────────────────────────────────────────────────┘
    ↓
doJobReport() - 序列化响应
    ↓
返回 JobResp（Success/Bury/Release）
    ↓
Processor ACK/Bury/Release 消息
```

---

## 四、依赖关系

### 依赖方向（从上到下）

```
cmd/worker/main.go
    ↓
internal/worker/manager.go
    ↓ (注入依赖)
internal/domains/processor.go
    ↓ (路由)
internal/domains/handlers/order/diagnose/handler.go
    ↓ (调用业务层)
internal/business/diagnosis_service.go
    ↓
internal/business/composite_handler.go
    ├─ internal/business/shipping_calculator.go
    └─ internal/business/anomaly_checker.go
    ↓
pkg/infra/mysql/order_dao.go
pkg/infra/redis/pubsub.go
```

### 依赖注入流程

```
Manager.NewManagerInstance()
  ├─ 初始化 OrderDAO（MySQL 连接）
  ├─ 初始化 RedisPubSub（Redis 连接）
  └─ 创建 DiagnosisService（注入 DAO 和 PubSub）

Manager.loadWorkers()
  └─ GetProcess(logger, diagnosisService)
      └─ 返回 lmstfyx.Proc 闭包（捕获 diagnosisService）

Processor.loop()
  └─ 调用 lmstfyx.Proc(ctx, job)
      └─ GetProcess 内部将 diagnosisService 注入到 Context

DiagnoseHandler.GetProcess()
  └─ 从 Context 获取 diagnosisService
      └─ 调用 diagnosisService.ExecuteDiagnosis()
```

---

## 五、为什么这样设计？

### 1. 框架与业务分离

**问题**：如果框架层直接耦合业务逻辑，难以复用和测试。

**解决**：
- 框架层（`framework/`）完全不知道业务，只负责消息拉取和分发
- 通过注入 `lmstfyx.Proc` 函数实现解耦
- 框架层可以独立测试、独立复用

### 2. 应用层与业务层分离

**问题**：如果 Handler 直接包含业务逻辑，难以单元测试和复用。

**解决**：
- 应用层（`domains/handlers/`）只负责路由和协调
- 业务层（`business/`）包含核心业务逻辑，可独立测试
- DiagnoseHandler 通过 DiagnosisService 调用业务逻辑

### 3. 依赖注入

**问题**：如果 Handler 直接创建依赖（OrderDAO、Redis），难以测试和替换。

**解决**：
- Manager 在启动时初始化所有依赖
- 通过 Context 传递依赖到 Handler
- 支持 Fallback 模式（未注入时使用 Mock）

### 4. 单一职责原则

每层只负责自己的职责：
- **框架层**：消息拉取 + 分发 + 优雅退出
- **Worker 层**：生命周期管理 + 依赖注入
- **应用层**：路由 + 协调 + 错误处理
- **业务层**：核心业务逻辑
- **基础设施层**：数据访问 + 外部服务

---

## 六、扩展指南

### 1. 添加新的诊断类型

在 `internal/business/` 添加新的业务逻辑：

```go
// internal/business/compliance_checker.go
type ComplianceChecker struct {}

func (c *ComplianceChecker) Check(ctx context.Context, orderData map[string]interface{}) (*model.ComplianceResult, error) {
    // 实现合规检查逻辑
}
```

在 `CompositeHandler` 中集成：

```go
func (h *CompositeHandler) Diagnose(ctx context.Context, input *DiagnoseInput) (*model.DiagnosisResultData, error) {
    items := make([]model.DiagnosisItem, 0, 3)

    // 原有诊断
    items = append(items, h.diagnoseShipping(ctx, input.OrderID))
    items = append(items, h.diagnoseAnomaly(ctx, input))

    // 新增合规检查
    items = append(items, h.diagnoseCompliance(ctx, input))

    return &model.DiagnosisResultData{Items: items}, nil
}
```

### 2. 添加新的 Handler

在 `internal/domains/handlers/` 创建新的 Handler：

```go
// internal/domains/handlers/order/refund/handler.go
type RefundHandler struct {
    ctx     context.Context
    meta    *job.Meta
    payload *RefundPayload
}

func (h *RefundHandler) GetProcess() *response.Response {
    // 实现退款逻辑
}
```

在 `handler_map.go` 注册：

```go
var HandlerMap = map[string]common.HandlerServProc{
    "order_diagnose": diagnose.NewDiagnoseHandler,
    "order_refund":   refund.NewRefundHandler, // 新增
}
```

### 3. 更换数据库

在 `pkg/infra/mysql/` 替换实现：

```go
// pkg/infra/postgres/order_dao.go
type OrderDAO struct {
    db *pgx.Conn // 使用 PostgreSQL
}

// 实现相同的接口
func (dao *OrderDAO) UpdateDiagnosisResult(...) error {
    // PostgreSQL 实现
}
```

在 `Manager` 中切换：

```go
// orderDAO, err := mysql.NewOrderDAO(cfg.MySQL.DSN)
orderDAO, err := postgres.NewOrderDAO(cfg.Postgres.DSN)
```

---

## 七、架构优势

✅ **清晰的职责划分**：每层只做一件事，易于理解和维护

✅ **高度可测试性**：业务层可独立测试，框架层可 Mock

✅ **易于扩展**：添加新功能只需修改少量代码

✅ **技术无关性**：业务层不依赖具体技术栈

✅ **依赖注入**：便于替换实现和测试

✅ **生产就绪**：经过充分测试和文档化

---

**版本**: v1.0.0
**最后更新**: 2025-12-23
