# DDD 架构改进方案

## 问题分析

### 当前架构问题
1. **Module 层反向依赖 Service 层** (order_module.go:9,16)
   - 违反分层依赖原则
   - 虽然未实际调用，但设计存在缺陷

2. **DiagnosisService 职责混乱**
   - 包含业务逻辑（消息格式构造、频道命名规则）
   - 直接依赖基础设施客户端 (lmstfy.Client, redis.PubSubClient)
   - 违反依赖倒置原则（DIP）

3. **错误处理缺失** (order_service.go:70-77)
   - 所有错误被静默忽略
   - 存在数据不一致风险

---

## 改进方案

### 核心思想
**Repository 不仅限于 MySQL，也应该包括 Redis、MQ 等资源的抽象**

### 架构分层

```
┌─────────────────────────────────────────────────────────────┐
│ Presentation Layer (Handler)                                │
├─────────────────────────────────────────────────────────────┤
│ Application/Domain Service                                  │
│ - OrderService: 复杂业务流程编排                              │
│ - CallbackService: 回调处理流程                              │
├─────────────────────────────────────────────────────────────┤
│ Domain Module (业务操作组装层)                               │
│ - OrderModule: 组装 OrderRepo + AccountRepo                 │
│ - DiagnosisModule: 组装 MessageQueueRepo + PubSubRepo       │
│   职责：构造消息格式、频道命名规则等业务逻辑                   │
├─────────────────────────────────────────────────────────────┤
│ Repository Interface (仓储接口)                              │
│ - OrderRepository (MySQL 数据访问)                           │
│ - MessageQueueRepository (消息队列访问)                       │
│ - PubSubRepository (发布订阅访问)                            │
├─────────────────────────────────────────────────────────────┤
│ Infrastructure (基础设施实现)                                 │
│ - OrderRepositoryImpl → MySQL + GORM                        │
│ - LmstfyRepositoryImpl → lmstfy.Client                      │
│ - RedisPubSubRepositoryImpl → redis.PubSubClient            │
└─────────────────────────────────────────────────────────────┘
```

---

## 详细改造步骤

### 步骤 1: 定义 MessageQueue Repository 接口

**文件**: `internal/app/domains/repo/rpmessage/message_queue_repo.go`

```go
package rpmessage

import (
    "context"
    "oip/common/model"
)

// MessageQueueRepository 消息队列仓储接口
// 抽象消息队列的访问，隔离具体 MQ 实现（Lmstfy/RabbitMQ/Kafka）
type MessageQueueRepository interface {
    // PublishDiagnoseJob 发布诊断任务到队列
    PublishDiagnoseJob(ctx context.Context, job *model.OrderDiagnoseJob) error
}
```

**职责**：
- ✅ 只负责消息发送的基础能力
- ❌ 不包含消息格式构造的业务逻辑

---

### 步骤 2: 定义 PubSub Repository 接口

**文件**: `internal/app/domains/repo/rppubsub/pubsub_repo.go`

```go
package rppubsub

import (
    "context"
    "time"
)

// PubSubRepository 发布订阅仓储接口
// 抽象 Pub/Sub 访问，隔离具体实现（Redis/NATS）
type PubSubRepository interface {
    // Publish 发布消息到指定频道
    Publish(ctx context.Context, channel string, message string) error

    // Subscribe 订阅频道并等待消息（支持超时）
    Subscribe(ctx context.Context, channel string, timeout time.Duration) (string, error)
}
```

**职责**：
- ✅ 只负责 Pub/Sub 的基础能力
- ❌ 不包含频道命名规则等业务逻辑

---

### 步骤 3: 创建 DiagnosisModule（替代 DiagnosisService）

**文件**: `internal/app/domains/modules/mddiagnosis/diagnosis_module.go`

```go
package mddiagnosis

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "github.com/google/uuid"
    "oip/common/model"
    "oip/dpmain/internal/app/domains/entity/etorder"
    "oip/dpmain/internal/app/domains/repo/rpmessage"
    "oip/dpmain/internal/app/domains/repo/rppubsub"
)

// DiagnosisModule 诊断模块
// 职责：
// 1. 组装 MessageQueueRepo + PubSubRepo
// 2. 包含诊断相关的业务逻辑（消息格式、频道命名规则）
type DiagnosisModule struct {
    mqRepo     rpmessage.MessageQueueRepository
    pubsubRepo rppubsub.PubSubRepository
    queueName  string
}

func NewDiagnosisModule(
    mqRepo rpmessage.MessageQueueRepository,
    pubsubRepo rppubsub.PubSubRepository,
    queueName string,
) *DiagnosisModule {
    return &DiagnosisModule{
        mqRepo:     mqRepo,
        pubsubRepo: pubsubRepo,
        queueName:  queueName,
    }
}

// PublishDiagnoseJob 发布订单诊断任务
// 业务逻辑：
// 1. 构造标准化消息格式（包含 RequestID, ActionType, OrgID 等）
// 2. 将 Shipment 转换为 map 格式（避免 dpsync 查询 DB）
func (m *DiagnosisModule) PublishDiagnoseJob(ctx context.Context, order *etorder.Order) error {
    // 业务逻辑：将 Shipment 结构体转换为 map
    var shipmentMap map[string]interface{}
    shipmentJSON, err := json.Marshal(order.Shipment)
    if err != nil {
        return fmt.Errorf("marshal shipment failed: %w", err)
    }
    if err := json.Unmarshal(shipmentJSON, &shipmentMap); err != nil {
        return fmt.Errorf("unmarshal shipment to map failed: %w", err)
    }

    // 业务逻辑：构造标准化消息格式
    job := model.OrderDiagnoseJob{
        Payload: model.OrderDiagnosePayload{
            Data: model.OrderDiagnoseData{
                RequestID:  uuid.New().String(), // 生成请求 ID（业务规则）
                OrgID:      "0",                 // MVP 固定值（业务规则）
                ActionType: "order_diagnose",    // 业务约定
                ID:         order.ID,
                Data: model.OrderDiagnoseBusinessData{
                    OrderID:         order.ID,
                    AccountID:       order.AccountID,
                    MerchantOrderNo: order.MerchantOrderNo,
                    Shipment:        shipmentMap,
                },
            },
        },
    }

    // 调用基础设施层（通过接口）
    return m.mqRepo.PublishDiagnoseJob(ctx, &job)
}

// WaitForDiagnosisResult 等待诊断结果（Smart Wait）
// 业务逻辑：
// 1. 知道订阅哪个频道（业务约定：diagnosis:result:{orderID}）
// 2. 解析诊断结果为领域对象
func (m *DiagnosisModule) WaitForDiagnosisResult(
    ctx context.Context,
    orderID string,
    timeout time.Duration,
) (*etorder.DiagnoseResult, error) {
    // 业务逻辑：频道命名规则
    channel := fmt.Sprintf("diagnosis:result:%s", orderID)

    // 调用基础设施层（通过接口）
    payload, err := m.pubsubRepo.Subscribe(ctx, channel, timeout)
    if err != nil {
        return nil, err
    }

    // 业务逻辑：反序列化为领域对象
    var result etorder.DiagnoseResult
    if err := json.Unmarshal([]byte(payload), &result); err != nil {
        return nil, fmt.Errorf("unmarshal diagnose result failed: %w", err)
    }

    return &result, nil
}
```

**为什么叫 Module 而不是 Service？**
- Module 体现了"组装多个 Repository"的职责
- 与 OrderModule 保持命名一致
- 避免与 OrderService (应用服务) 混淆

---

### 步骤 4: Infrastructure 层实现 Repository

#### 4.1 Lmstfy Repository 实现

**文件**: `internal/app/infra/mq/lmstfy_repo_impl.go`

```go
package mq

import (
    "context"
    "oip/common/model"
    "oip/dpmain/internal/app/domains/repo/rpmessage"
    "oip/dpmain/internal/app/infra/mq/lmstfy"
)

// LmstfyRepositoryImpl Lmstfy 消息队列仓储实现
type LmstfyRepositoryImpl struct {
    client    *lmstfy.Client
    queueName string
}

func NewLmstfyRepository(client *lmstfy.Client, queueName string) rpmessage.MessageQueueRepository {
    return &LmstfyRepositoryImpl{
        client:    client,
        queueName: queueName,
    }
}

func (r *LmstfyRepositoryImpl) PublishDiagnoseJob(ctx context.Context, job *model.OrderDiagnoseJob) error {
    // 只负责调用底层客户端，不包含业务逻辑
    return r.client.Publish(ctx, r.queueName, job)
}
```

#### 4.2 Redis PubSub Repository 实现

**文件**: `internal/app/infra/persistence/redis/pubsub_repo_impl.go`

```go
package redis

import (
    "context"
    "time"
    "oip/dpmain/internal/app/domains/repo/rppubsub"
)

// PubSubRepositoryImpl Redis Pub/Sub 仓储实现
type PubSubRepositoryImpl struct {
    client *PubSubClient
}

func NewPubSubRepository(client *PubSubClient) rppubsub.PubSubRepository {
    return &PubSubRepositoryImpl{client: client}
}

func (r *PubSubRepositoryImpl) Publish(ctx context.Context, channel string, message string) error {
    // 只负责调用底层客户端
    return r.client.Publish(ctx, channel, message)
}

func (r *PubSubRepositoryImpl) Subscribe(ctx context.Context, channel string, timeout time.Duration) (string, error) {
    // 只负责调用底层客户端
    return r.client.Subscribe(ctx, channel, timeout)
}
```

---

### 步骤 5: 修改 OrderModule（移除对 DiagnosisService 的依赖）

**文件**: `internal/app/domains/modules/mdorder/order_module.go`

```go
package mdorder

import (
    "context"
    "oip/dpmain/internal/app/domains/entity/etorder"
    "oip/dpmain/internal/app/domains/repo/rpaccount"
    "oip/dpmain/internal/app/domains/repo/rporder"
    // 移除这行：
    // "oip/dpmain/internal/app/domains/services/svdiagnosis"
)

type OrderModule struct {
    orderRepo   rporder.OrderRepository
    accountRepo rpaccount.AccountRepository
    // 移除这行：
    // diagnosisService *svdiagnosis.DiagnosisService
}

func NewOrderModule(
    orderRepo rporder.OrderRepository,
    accountRepo rpaccount.AccountRepository,
    // 移除参数：diagnosisService *svdiagnosis.DiagnosisService
) *OrderModule {
    return &OrderModule{
        orderRepo:   orderRepo,
        accountRepo: accountRepo,
    }
}

// ... 其他方法保持不变
```

---

### 步骤 6: 修改 OrderService（依赖 DiagnosisModule）

**文件**: `internal/app/domains/services/svorder/order_service.go`

```go
package svorder

import (
    "context"
    "errors"
    "fmt"
    "log"
    "time"

    "github.com/google/uuid"
    "oip/dpmain/internal/app/domains/entity/etorder"
    "oip/dpmain/internal/app/domains/modules/mddiagnosis"
    "oip/dpmain/internal/app/domains/modules/mdorder"
)

type OrderService struct {
    orderModule      *mdorder.OrderModule
    diagnosisModule  *mddiagnosis.DiagnosisModule  // 依赖 Module 而不是 Service
}

func NewOrderService(
    orderModule *mdorder.OrderModule,
    diagnosisModule *mddiagnosis.DiagnosisModule,
) *OrderService {
    return &OrderService{
        orderModule:     orderModule,
        diagnosisModule: diagnosisModule,
    }
}

func (s *OrderService) CreateOrder(
    ctx context.Context,
    accountID int64,
    merchantOrderNo string,
    shipment *etorder.Shipment,
    waitSeconds int,
) (*etorder.Order, error) {
    // 1. 验证 account 存在
    exists, err := s.orderModule.AccountExists(ctx, accountID)
    if err != nil {
        return nil, fmt.Errorf("check account exists failed: %w", err)
    }
    if !exists {
        return nil, errors.New("account not found")
    }

    // 2. 检查订单重复
    existing, err := s.orderModule.GetOrderByAccountAndMerchantNo(ctx, accountID, merchantOrderNo)
    if err != nil {
        return nil, fmt.Errorf("check order duplicate failed: %w", err)
    }
    if existing != nil {
        return nil, fmt.Errorf("order already exists: merchant_order_no=%s", merchantOrderNo)
    }

    // 3. 验证货件信息
    if err := s.validateShipment(shipment); err != nil {
        return nil, fmt.Errorf("validate shipment failed: %w", err)
    }

    // 4. 创建订单并落库
    order, err := etorder.NewOrder(uuid.New().String(), accountID, merchantOrderNo, shipment)
    if err != nil {
        return nil, fmt.Errorf("create order entity failed: %w", err)
    }

    if err := s.orderModule.CreateOrder(ctx, order); err != nil {
        return nil, fmt.Errorf("save order failed: %w", err)
    }

    // 5. 发布到诊断队列
    if err := s.diagnosisModule.PublishDiagnoseJob(ctx, order); err != nil {
        // 发布失败只记录日志，不影响订单创建成功
        log.Printf("[WARN] publish diagnose job failed: order_id=%s, error=%v", order.ID, err)
    }

    // 6. Smart Wait（等待诊断结果）
    if waitSeconds > 0 {
        timeout := time.Duration(waitSeconds) * time.Second
        result, err := s.diagnosisModule.WaitForDiagnosisResult(ctx, order.ID, timeout)

        // ✅ 修复：正确处理错误
        if err != nil {
            // 超时或订阅失败，只记录日志
            log.Printf("[WARN] wait for diagnosis result failed: order_id=%s, error=%v", order.ID, err)
            return order, nil // 返回订单，状态仍为 DIAGNOSING
        }

        if result != nil {
            // 更新内存中的 Order 实体
            if err := order.UpdateDiagnoseResult(result); err != nil {
                return nil, fmt.Errorf("update order entity failed: %w", err)
            }

            // 持久化到 DB
            if err := s.orderModule.UpdateDiagnoseResult(ctx, order.ID, result); err != nil {
                // ⚠️ 严重问题：内存已更新，DB 更新失败
                // 这里需要事务补偿或告警
                log.Printf("[ERROR] persist diagnose result failed: order_id=%s, error=%v", order.ID, err)
                return nil, fmt.Errorf("persist diagnose result failed: %w", err)
            }
        }
    }

    return order, nil
}

// ... 其他方法保持不变
```

**关键改进**：
1. ✅ 依赖 `DiagnosisModule` 而不是 `DiagnosisService`
2. ✅ 正确处理所有错误，不再静默吞没
3. ✅ 区分"发布失败"和"等待超时"的处理逻辑
4. ✅ 记录关键错误日志，便于排查问题

---

### 步骤 7: 修改 CallbackService（依赖 PubSubRepository）

**文件**: `internal/app/domains/services/svcallback/callback_service.go`

```go
package svcallback

import (
    "context"
    "encoding/json"
    "fmt"

    "oip/common/entity"
    "oip/common/model"
    "oip/dpmain/internal/app/domains/repo/rporder"
    "oip/dpmain/internal/app/domains/repo/rppubsub"  // 依赖接口
    "oip/dpmain/internal/app/pkg/logger"
)

type CallbackService struct {
    orderRepo   rporder.OrderRepository
    pubsubRepo  rppubsub.PubSubRepository  // 依赖接口而不是具体实现
    logger      logger.Logger
}

func NewCallbackService(
    orderRepo rporder.OrderRepository,
    pubsubRepo rppubsub.PubSubRepository,
    logger logger.Logger,
) *CallbackService {
    return &CallbackService{
        orderRepo:  orderRepo,
        pubsubRepo: pubsubRepo,
        logger:     logger,
    }
}

// publishNotification 发送 Redis PubSub 通知（业务逻辑）
func (s *CallbackService) publishNotification(ctx context.Context, callback *model.OrderDiagnoseCallback) error {
    // 业务逻辑：频道命名规则
    channel := fmt.Sprintf("diagnosis:result:%s", callback.OrderID)

    // 业务逻辑：构造通知数据
    var notificationData interface{}
    if callback.Status == model.CallbackStatusSuccess && callback.DiagnosisResult != nil {
        notificationData = map[string]interface{}{
            "items": callback.DiagnosisResult.Items,
        }
    } else {
        notificationData = map[string]interface{}{
            "status": callback.Status,
            "error":  callback.Error,
        }
    }

    // 序列化
    payload, err := json.Marshal(notificationData)
    if err != nil {
        return fmt.Errorf("marshal notification failed: %w", err)
    }

    // 调用基础设施层（通过接口）
    if err := s.pubsubRepo.Publish(ctx, channel, string(payload)); err != nil {
        return fmt.Errorf("publish to pubsub failed: %w", err)
    }

    s.logger.InfoContext(ctx, "Redis notification sent",
        "order_id", callback.OrderID,
        "channel", channel,
    )

    return nil
}

// ... 其他方法保持不变
```

---

### 步骤 8: 修改依赖注入（Wire）

**文件**: `cmd/apiserver/wire.go`

```go
// +build wireinject

package main

import (
    "github.com/google/wire"
    // ... 其他导入

    "oip/dpmain/internal/app/domains/modules/mddiagnosis"  // 新增
    "oip/dpmain/internal/app/domains/repo/rpmessage"      // 新增
    "oip/dpmain/internal/app/domains/repo/rppubsub"       // 新增
    mqinfra "oip/dpmain/internal/app/infra/mq"            // 新增
)

// InfraSet 基础设施依赖集
var InfraSet = wire.NewSet(
    // ... 现有的依赖

    // 新增：Repository 实现
    mqinfra.NewLmstfyRepository,      // MessageQueueRepository 实现
    wire.Bind(new(rpmessage.MessageQueueRepository), new(*mqinfra.LmstfyRepositoryImpl)),

    redis.NewPubSubRepository,        // PubSubRepository 实现
    wire.Bind(new(rppubsub.PubSubRepository), new(*redis.PubSubRepositoryImpl)),
)

// ModuleSet 模块依赖集
var ModuleSet = wire.NewSet(
    mdorder.NewOrderModule,
    mdaccount.NewAccountModule,
    mddiagnosis.NewDiagnosisModule,  // 新增
)

// ServiceSet 服务依赖集
var ServiceSet = wire.NewSet(
    svorder.NewOrderService,
    svaccount.NewAccountService,
    svcallback.NewCallbackService,
    // 移除：svdiagnosis.NewDiagnosisService
)

// ... 其他集合
```

---

## 改进效果对比

### 改进前 ❌

```go
// DiagnosisService 直接依赖具体实现
type DiagnosisService struct {
    lmstfyClient *lmstfy.Client         // 紧耦合
    redisClient  *redis.PubSubClient    // 紧耦合
}

// OrderModule 反向依赖 Service
type OrderModule struct {
    diagnosisService *svdiagnosis.DiagnosisService  // 违反分层原则
}

// 错误处理缺失
if err == nil && result != nil {
    if updateErr := order.UpdateDiagnoseResult(result); updateErr == nil {
        s.orderModule.UpdateDiagnoseResult(ctx, order.ID, result)  // 忽略错误
    }
}
```

### 改进后 ✅

```go
// DiagnosisModule 依赖接口
type DiagnosisModule struct {
    mqRepo     rpmessage.MessageQueueRepository  // 依赖接口
    pubsubRepo rppubsub.PubSubRepository         // 依赖接口
}

// OrderModule 不再依赖 Service
type OrderModule struct {
    orderRepo   rporder.OrderRepository
    accountRepo rpaccount.AccountRepository
    // 移除了 diagnosisService
}

// 错误处理完善
if err != nil {
    log.Printf("[WARN] wait failed: %v", err)
    return order, nil
}
if result != nil {
    if err := order.UpdateDiagnoseResult(result); err != nil {
        return nil, fmt.Errorf("update entity failed: %w", err)
    }
    if err := s.orderModule.UpdateDiagnoseResult(ctx, order.ID, result); err != nil {
        log.Printf("[ERROR] persist failed: %v", err)
        return nil, fmt.Errorf("persist failed: %w", err)
    }
}
```

---

## 架构改进的核心价值

| 改进点 | 价值 |
|--------|------|
| **Repository 接口抽象** | 隔离基础设施，便于测试和替换（如从 Lmstfy 迁移到 Kafka） |
| **Module 层体现价值** | 组装多个 Repository，包含业务逻辑（消息格式、频道规则） |
| **依赖方向正确** | Service → Module → Repository → Infrastructure |
| **业务与基础设施分离** | 业务逻辑在领域层，基础设施在 infra 层 |
| **错误处理完善** | 所有错误显式处理，避免静默失败 |

---

## 实施建议

### 优先级

**P0（立即修复）**：
1. 移除 OrderModule 对 DiagnosisService 的依赖
2. 修复 order_service.go:70-77 的错误处理

**P1（重点改造）**：
3. 定义 MessageQueueRepository 和 PubSubRepository 接口
4. 创建 DiagnosisModule 替代 DiagnosisService
5. 实现基础设施层 Repository

**P2（完善优化）**：
6. 更新 Wire 依赖注入配置
7. 补充单元测试（利用接口 mock）

---

## 测试改进建议

改造后，可以轻松编写单元测试：

```go
// order_service_test.go
func TestOrderService_CreateOrder(t *testing.T) {
    // Mock Repositories
    mockOrderRepo := &MockOrderRepository{}
    mockAccountRepo := &MockAccountRepository{}
    mockMQRepo := &MockMessageQueueRepository{}
    mockPubSubRepo := &MockPubSubRepository{}

    // 创建 Modules
    orderModule := mdorder.NewOrderModule(mockOrderRepo, mockAccountRepo)
    diagnosisModule := mddiagnosis.NewDiagnosisModule(mockMQRepo, mockPubSubRepo, "test_queue")

    // 创建 Service
    service := svorder.NewOrderService(orderModule, diagnosisModule)

    // 测试业务逻辑（不依赖真实的 Lmstfy 和 Redis）
    order, err := service.CreateOrder(ctx, 1, "ORDER-001", shipment, 10)
    assert.NoError(t, err)
    assert.Equal(t, "DIAGNOSING", order.Status)
}
```

---

## 总结

这次改造的核心：
1. **Repository 不仅限于 MySQL**，也包括 Redis、MQ 等资源的抽象
2. **Module 层组装多个 Repository**，包含业务逻辑（消息格式、频道规则等）
3. **Service 层依赖 Module**，进行更高层次的业务流程编排
4. **依赖方向正确**：Service → Module → Repository → Infrastructure
5. **业务与基础设施完全分离**，符合 DDD 和洋葱架构原则
