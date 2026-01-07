# OIP 代码规范

> 保持代码一致性和可维护性

## 代码风格

### 1. 命名规范

#### 包名（Package）
```go
// ✅ 小写，单数，简短
package order
package user
package redis

// ❌ 避免
package orderPackage  // 不要加 Package 后缀
package Orders        // 不要复数
```

#### 文件名
```go
// ✅ 小写，下划线分隔
order_service.go
user_handler.go
redis_client.go

// ❌ 避免
OrderService.go
userHandler.go
```

#### 变量名
```go
// ✅ 驼峰命名，首字母小写（私有）或大写（公开）
var userID string
var orderNumber string

// ❌ 避免
var user_id string     // 不要用下划线
var UserID string      // 私有变量不要大写开头
```

#### 常量
```go
// ✅ 全大写，下划线分隔（仅导出常量）
const (
    OrderStatusCreated  = "CREATED"
    OrderStatusCancelled = "CANCELLED"
)

// ✅ 驼峰命名（私有常量）
const (
    defaultTimeout = 10 * time.Second
    maxRetries     = 3
)
```

#### 接口命名
```go
// ✅ 单方法接口：动词 + er
type Reader interface {
    Read(p []byte) (n int, err error)
}

type Writer interface {
    Write(p []byte) (n int, err error)
}

// ✅ 多方法接口：名词
type OrderRepository interface {
    Create(ctx context.Context, order *Order) error
    GetByID(ctx context.Context, id string) (*Order, error)
}
```

### 2. 注释规范

#### 包注释
```go
// Package order provides order management functionality.
// It includes order creation, query, and status management.
package order
```

#### 函数注释
```go
// CreateOrder creates a new order with the given payload.
// It validates the payload, persists the order, and returns the order ID.
//
// Example:
//   order, err := CreateOrder(ctx, payload)
//   if err != nil {
//       log.Error(err)
//   }
func CreateOrder(ctx context.Context, payload *CreateOrderPayload) (*Order, error) {
    // ...
}
```

#### 类型注释
```go
// Order represents an order entity with its associated metadata.
type Order struct {
    ID          string    `json:"id"`           // Order unique identifier
    OrderNumber string    `json:"order_number"` // Business order number
    Status      string    `json:"status"`       // Order status
}
```

### 3. 错误处理

#### 错误定义
```go
// ✅ 使用 errors.New 定义错误
var (
    ErrOrderNotFound      = errors.New("order not found")
    ErrInvalidOrderStatus = errors.New("invalid order status")
)

// ✅ 使用 errors.Wrap 保留堆栈
func GetOrder(ctx context.Context, id string) (*Order, error) {
    order, err := orderDAO.GetByID(ctx, id)
    if err != nil {
        return nil, errors.Wrap(err, "failed to get order from DAO")
    }
    return order, nil
}
```

#### 错误检查
```go
// ✅ 先检查错误
order, err := GetOrder(ctx, id)
if err != nil {
    return nil, err
}

// 使用 order...

// ❌ 避免忽略错误
order, _ := GetOrder(ctx, id) // 不要忽略错误
```

### 4. 日志规范

```go
// ✅ 使用结构化日志
logger.Infof(ctx, "order created: order_id=%s, account_id=%d", order.ID, order.AccountID)

logger.Errorf(ctx, "failed to create order: err=%v, payload=%+v", err, payload)

// ❌ 避免
log.Println("order created", order.ID) // 不要用 Println
```

## 代码组织

### 1. 分层架构

```
domains/               # 领域层（核心业务逻辑）
  ├── entity/         # 聚合根和值对象
  ├── apimodel/       # API 模型（DTO）
  ├── modules/        # 业务编排
  ├── repo/           # 仓储接口
  └── services/       # 领域服务

infra/                # 基础设施层
  ├── persistence/    # 持久化实现
  └── mq/             # 消息队列

server/               # 服务器层
  ├── handlers/       # HTTP 处理器
  ├── routers/        # 路由
  └── middlewares/    # 中间件

pkg/                  # 工具包
  ├── errorx/         # 错误处理
  ├── ginx/           # Gin 扩展
  └── logger/         # 日志
```

### 2. 文件组织

#### 同一功能的代码组织在一起
```
domains/entity/etorder/
├── order.go          # Order 聚合根
├── order_item.go     # OrderItem 实体
└── types.go          # 枚举和常量
```

#### 测试文件与源文件同目录
```
order_service.go
order_service_test.go
```

### 3. 依赖管理

#### 依赖方向
```
server (HTTP 适配)
  ↓
modules (业务编排)
  ↓
services (领域服务) + repo (仓储接口)
  ↑ 实现
infra (基础设施)
```

**原则**：依赖指向领域层，领域层不依赖外部

## 代码实践

### 1. 使用 Context

```go
// ✅ 所有公开方法都接收 context.Context
func CreateOrder(ctx context.Context, payload *CreateOrderPayload) (*Order, error) {
    // ...
}

// ✅ 传递 context
func (s *OrderService) CreateOrder(ctx context.Context, payload *CreateOrderPayload) (*Order, error) {
    order, err := s.orderModule.CreateOrder(ctx, payload)
    if err != nil {
        return nil, err
    }
    return order, nil
}
```

### 2. 使用 defer

```go
// ✅ 使用 defer 释放资源
func CreateOrder(ctx context.Context, payload *CreateOrderPayload) error {
    tx := db.Begin()
    defer func() {
        if err != nil {
            tx.Rollback()
        } else {
            tx.Commit()
        }
    }()

    // ...
}

// ✅ 使用 defer 关闭订阅
sub := redis.Subscribe(channel)
defer sub.Close()
```

### 3. 使用 Interface

```go
// ✅ 定义接口，面向接口编程
type OrderRepository interface {
    Create(ctx context.Context, order *Order) error
    GetByID(ctx context.Context, id string) (*Order, error)
}

// ✅ 依赖接口，不依赖实现
type OrderService struct {
    orderRepo OrderRepository // 接口
}
```

### 4. 避免全局变量

```go
// ❌ 避免
var globalDB *gorm.DB

func init() {
    globalDB, _ = gorm.Open(...)
}

// ✅ 使用依赖注入
type OrderDAO struct {
    db *gorm.DB
}

func NewOrderDAO(db *gorm.DB) *OrderDAO {
    return &OrderDAO{db: db}
}
```

### 5. 配置管理

```go
// ✅ 使用配置结构体
type Config struct {
    Server ServerConfig
    MySQL  MySQLConfig
    Redis  RedisConfig
}

// ✅ 从文件加载配置
func LoadConfig() (*Config, error) {
    viper.SetConfigFile("config.yaml")
    viper.ReadInConfig()

    var cfg Config
    viper.Unmarshal(&cfg)
    return &cfg, nil
}

// ❌ 避免硬编码
const ServerPort = 7777  // 硬编码
```

## 测试规范

### 1. 测试命名

```go
// ✅ Test + 函数名
func TestCreateOrder(t *testing.T) {
    // ...
}

// ✅ Test + 函数名 + 场景
func TestCreateOrder_Success(t *testing.T) {
    // ...
}

func TestCreateOrder_OrderAlreadyExists(t *testing.T) {
    // ...
}
```

### 2. 表驱动测试

```go
func TestCalculateTotal(t *testing.T) {
    tests := []struct {
        name     string
        input    []OrderItem
        expected float64
    }{
        {"empty items", []OrderItem{}, 0.0},
        {"single item", []OrderItem{{Quantity: 2, Price: 10.0}}, 20.0},
        {"multiple items", []OrderItem{{Quantity: 2, Price: 10.0}, {Quantity: 3, Price: 5.0}}, 35.0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := calculateTotal(tt.input)
            if got != tt.expected {
                t.Errorf("expected %v, got %v", tt.expected, got)
            }
        })
    }
}
```

### 3. Mock 使用

```go
// ✅ 使用接口 Mock
type MockOrderRepository struct {
    CreateFunc  func(ctx context.Context, order *Order) error
    GetByIDFunc func(ctx context.Context, id string) (*Order, error)
}

func (m *MockOrderRepository) Create(ctx context.Context, order *Order) error {
    if m.CreateFunc != nil {
        return m.CreateFunc(ctx, order)
    }
    return nil
}
```

## 性能优化

### 1. 避免不必要的内存分配

```go
// ✅ 预分配切片容量
items := make([]OrderItem, 0, 10)

// ❌ 避免频繁扩容
items := []OrderItem{}
for i := 0; i < 10; i++ {
    items = append(items, item) // 频繁扩容
}
```

### 2. 使用 sync.Pool

```go
var bufferPool = sync.Pool{
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

func ProcessData(data []byte) {
    buf := bufferPool.Get().(*bytes.Buffer)
    defer bufferPool.Put(buf)

    buf.Reset()
    buf.Write(data)
    // ...
}
```

### 3. 批量操作

```go
// ✅ 批量插入
func BatchCreateOrders(ctx context.Context, orders []*Order) error {
    return db.Create(orders).Error
}

// ❌ 避免循环单条插入
for _, order := range orders {
    db.Create(order)
}
```

## Git 提交规范

### Commit Message 格式

```
<type>(<scope>): <subject>

<body>

<footer>
```

### Type 类型
- `feat`: 新功能
- `fix`: Bug 修复
- `refactor`: 重构
- `docs`: 文档
- `test`: 测试
- `chore`: 构建、工具

### 示例
```
feat(order): add smart wait mechanism

- Implement Redis Pub/Sub for order diagnosis notification
- Add 10s timeout for diagnosis result
- Return 3001 Processing when timeout

Closes #123
```

## 持续集成

### Pre-commit 检查
```bash
# 格式化
gofmt -w .

# 静态检查
go vet ./...

# 运行测试
go test ./... -v

# 依赖整理
go mod tidy
```

### Code Review 要点
- [ ] 代码符合分层架构
- [ ] 错误处理完整
- [ ] 有必要的注释
- [ ] 有单元测试
- [ ] 无硬编码配置
- [ ] 无循环依赖
