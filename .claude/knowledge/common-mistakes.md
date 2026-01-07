# 常见错误与修正方案

> 每次遇到错误后追加到这里，避免重复犯错

## 错误 1: 循环依赖

### 现象
```
package import cycle
```

### 错误示例
```go
// internal/service/order_service.go
import "oip/internal/handler"

// internal/handler/order_handler.go
import "oip/internal/service"
```

### 原因
- service 和 handler 互相import

### 修正方案
✅ 使用 interface 解耦，依赖倒置

```go
// internal/service/order_service.go
type OrderService interface {
    CreateOrder(ctx context.Context, req *CreateOrderRequest) error
}

// internal/handler/order_handler.go
type OrderHandler struct {
    orderService service.OrderService // 依赖接口
}
```

---

## 错误 2: 硬编码配置

### 现象
```go
// 代码中直接写死配置
conn, err := grpc.Dial("localhost:7777", ...)
```

### 错误示例
```go
func NewClient() *Client {
    return &Client{
        host: "localhost",
        port: 7777,
    }
}
```

### 原因
- 配置硬编码，难以适配不同环境

### 修正方案
✅ 从 config.yaml 读取配置

```go
// config/config.yaml
server:
  host: "localhost"
  port: 7777

// 代码中读取
func NewClient(cfg *Config) *Client {
    return &Client{
        host: cfg.Server.Host,
        port: cfg.Server.Port,
    }
}
```

---

## 错误 3: 忘记事务回滚

### 现象
数据库连接泄漏，事务未提交或回滚

### 错误示例
```go
func CreateOrder(ctx context.Context, order *Order) error {
    tx := db.Begin()

    if err := orderDAO.Create(tx, order); err != nil {
        return err // 忘记回滚
    }

    if err := itemDAO.Create(tx, order.Items); err != nil {
        return err // 忘记回滚
    }

    tx.Commit()
    return nil
}
```

### 原因
- 错误处理路径未回滚事务
- 容易导致连接泄漏

### 修正方案
✅ 使用 defer 自动回滚

```go
func CreateOrder(ctx context.Context, order *Order) error {
    tx := db.Begin()

    var err error
    defer func() {
        if err != nil {
            tx.Rollback()
        } else {
            tx.Commit()
        }
    }()

    if err = orderDAO.Create(tx, order); err != nil {
        return err
    }

    if err = itemDAO.Create(tx, order.Items); err != nil {
        return err
    }

    return nil
}
```

---

## 错误 4: 在 handler 层直接操作数据库

### 现象
handler 层违反分层原则，直接操作数据库

### 错误示例
```go
// handlers/order/create.go
func CreateOrder(c *gin.Context) {
    var req CreateOrderRequest
    c.BindJSON(&req)

    // ❌ handler 直接操作数据库
    order := &entity.Order{...}
    db.Create(order)

    c.JSON(200, order)
}
```

### 原因
- 违反分层架构
- 难以测试
- 业务逻辑分散

### 修正方案
✅ 调用模块/服务层

```go
// handlers/order/create.go
func CreateOrder(c *gin.Context) {
    var req CreateOrderRequest
    c.BindJSON(&req)

    // ✅ 调用模块
    order, err := orderModule.CreateOrder(ctx, &req)
    if err != nil {
        ginx.ResponseError(c, err)
        return
    }

    ginx.ResponseWithOK(c, order)
}
```

---

## 错误 5: 未处理 Redis 订阅泄漏

### 现象
Redis 连接数持续增长，最终耗尽连接池

### 错误示例
```go
func SmartWait(orderID string, timeout time.Duration) (*Result, error) {
    channel := fmt.Sprintf("diagnosis:result:%s", orderID)
    sub := redis.Subscribe(channel)
    // ❌ 忘记关闭订阅

    select {
    case msg := <-sub.Channel():
        return parseResult(msg), nil
    case <-time.After(timeout):
        return nil, ErrTimeout
    }
}
```

### 原因
- 订阅未关闭，连接泄漏

### 修正方案
✅ 使用 defer 关闭订阅

```go
func SmartWait(orderID string, timeout time.Duration) (*Result, error) {
    channel := fmt.Sprintf("diagnosis:result:%s", orderID)
    sub := redis.Subscribe(channel)
    defer sub.Close() // ✅ 确保关闭

    select {
    case msg := <-sub.Channel():
        return parseResult(msg), nil
    case <-time.After(timeout):
        return nil, ErrTimeout
    }
}
```

---

## 错误 6: 错误处理丢失堆栈信息

### 现象
错误日志只有错误消息，没有堆栈信息，难以定位问题

### 错误示例
```go
func CreateOrder(ctx context.Context, req *CreateOrderRequest) error {
    err := orderDAO.Create(ctx, order)
    if err != nil {
        return err // ❌ 直接返回，丢失堆栈
    }
    return nil
}
```

### 原因
- 直接返回 error，丢失调用栈

### 修正方案
✅ 使用 errors.Wrap 保留堆栈

```go
import "github.com/pkg/errors"

func CreateOrder(ctx context.Context, req *CreateOrderRequest) error {
    err := orderDAO.Create(ctx, order)
    if err != nil {
        return errors.Wrap(err, "failed to create order in DAO") // ✅ 保留堆栈
    }
    return nil
}
```

---

## 错误 7: Goroutine 泄漏

### 现象
Goroutine 数量持续增长，内存泄漏

### 错误示例
```go
func ProcessOrder(orderID string) {
    go func() {
        // ❌ 无限循环，goroutine 永不退出
        for {
            process(orderID)
            time.Sleep(1 * time.Second)
        }
    }()
}
```

### 原因
- Goroutine 没有退出机制

### 修正方案
✅ 使用 context 控制生命周期

```go
func ProcessOrder(ctx context.Context, orderID string) {
    go func() {
        ticker := time.NewTicker(1 * time.Second)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done(): // ✅ 监听 context 取消
                return
            case <-ticker.C:
                process(orderID)
            }
        }
    }()
}
```

---

## 错误 8: JSON 序列化/反序列化错误

### 现象
JSON 字段为空或解析失败

### 错误示例
```go
type Order struct {
    ID          string    // ❌ 没有 json tag
    OrderNumber string    // ❌ 没有 json tag
    Status      string    // ❌ 没有 json tag
}
```

### 原因
- 结构体字段未导出（小写开头）
- 缺少 json tag

### 修正方案
✅ 添加 json tag

```go
type Order struct {
    ID          string `json:"id"`
    OrderNumber string `json:"order_number"`
    Status      string `json:"status"`
}
```

---

## 错误 9: 未验证请求参数

### 现象
API 接收到非法参数，导致运行时错误

### 错误示例
```go
func CreateOrder(c *gin.Context) {
    var req CreateOrderRequest
    c.BindJSON(&req)
    // ❌ 未验证参数

    order := NewOrder(req)
    // ...
}
```

### 原因
- 缺少参数验证
- 业务逻辑假设参数合法

### 修正方案
✅ 使用 Gin 验证器

```go
type CreateOrderRequest struct {
    AccountID       int64     `json:"account_id" binding:"required"`
    MerchantOrderNo string    `json:"merchant_order_no" binding:"required,min=1,max=128"`
    Shipment        *Shipment `json:"shipment" binding:"required"`
}

func CreateOrder(c *gin.Context) {
    var req CreateOrderRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        ginx.ResponseError(c, err) // ✅ 验证失败返回错误
        return
    }
    // ...
}
```

---

## 错误 10: 未处理并发竞争

### 现象
数据不一致、panic（concurrent map writes）

### 错误示例
```go
type Cache struct {
    data map[string]string // ❌ 未加锁
}

func (c *Cache) Set(key, value string) {
    c.data[key] = value // ❌ 并发写入 panic
}
```

### 原因
- 多个 Goroutine 并发访问共享数据
- 未使用锁保护

### 修正方案
✅ 使用 sync.RWMutex

```go
type Cache struct {
    mu   sync.RWMutex
    data map[string]string
}

func (c *Cache) Set(key, value string) {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.data[key] = value
}

func (c *Cache) Get(key string) (string, bool) {
    c.mu.RLock()
    defer c.mu.RUnlock()
    val, ok := c.data[key]
    return val, ok
}
```

---

## 持续更新

每次遇到新的错误，请按以下格式追加：

### 错误 X: [错误名称]

**现象**: [描述问题表现]

**错误示例**: [代码示例]

**原因**: [分析原因]

**修正方案**: [正确写法]
