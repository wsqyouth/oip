# dpmain - OIP 同步 API 服务

## 架构说明

基于 **DDD（领域驱动设计）** 的生产级架构，采用 **单进程多 Goroutine** 设计。

### 🚀 单进程架构

**设计理念**：将 HTTP API Server 和 Callback Consumer 整合到同一个进程中。

```
┌─────────────────────────────────────┐
│         dpmain (单进程)             │
│  ┌────────────┐  ┌───────────────┐  │
│  │   HTTP     │  │   Callback    │  │
│  │   Server   │  │   Consumer    │  │
│  │ (Goroutine)│  │  (Goroutine)  │  │
│  └────────────┘  └───────────────┘  │
└─────────────────────────────────────┘
           ↓
   ┌───────────────────┐
   │  MySQL / Redis    │
   │   Lmstfy (MQ)     │
   └───────────────────┘
```

**优势**：
- ✅ 单一启动入口（`./bin/apiserver`）
- ✅ 共享数据库连接池和 Redis 连接
- ✅ 统一的日志输出和监控
- ✅ 简化部署和运维
- ✅ 优雅停机（协调 HTTP + Consumer）

### 目录结构

```
dpmain/
├── cmd/apiserver/              # 程序入口（单进程）
│   ├── main.go                 # 启动 HTTP + Consumer
│   ├── wire.go                 # Wire 依赖注入配置
│   └── wire_gen.go             # Wire 生成的代码
│
├── internal/app/
│   ├── domains/                # 【领域层】
│   │   ├── entity/             # 实体（纯领域对象）
│   │   │   ├── etorder/        # Order 聚合根
│   │   │   ├── etaccount/      # Account 实体
│   │   │   └── etprimitive/    # 基础类型
│   │   ├── apimodel/           # API 模型（DTO）
│   │   │   ├── request/        # 请求 DTO
│   │   │   └── response/       # 响应 DTO
│   │   ├── modules/            # 领域模块（业务编排）
│   │   │   ├── mdorder/        # Order 模块
│   │   │   └── mdaccount/      # Account 模块
│   │   ├── repo/               # 仓储接口（只定义）
│   │   │   ├── rporder/        # OrderRepository
│   │   │   └── rpaccount/      # AccountRepository
│   │   └── services/           # 领域服务（复杂逻辑）
│   │       ├── svorder/        # Order 服务
│   │       ├── svcallback/     # Callback 服务
│   │       └── svdiagnosis/    # Diagnosis 服务
│   │
│   ├── infra/                  # 【基础设施层】
│   │   ├── persistence/        # 持久化实现
│   │   │   ├── mysql/          # MySQL 仓储实现
│   │   │   └── redis/          # Redis Pub/Sub
│   │   └── mq/                 # 消息队列
│   │       └── lmstfy/         # Lmstfy 客户端
│   │
│   ├── consumer/               # 【消费者层】
│   │   └── callback_consumer.go # 回调消费者
│   │
│   ├── server/                 # 【服务器层】
│   │   ├── handlers/           # HTTP 处理器
│   │   │   ├── order/          # Order 处理器
│   │   │   └── account/        # Account 处理器
│   │   ├── routers/            # 路由配置
│   │   └── middlewares/        # 中间件
│   │
│   ├── pkg/                    # 【通用包】
│   │   ├── errorx/             # 错误处理
│   │   ├── ginx/               # Gin 扩展
│   │   └── logger/             # 日志
│   │
│   ├── config/                 # 配置管理
│   └── utils/                  # 工具函数
│
├── scripts/                    # 构建脚本
├── go.mod                      # 模块依赖
├── Makefile                    # 构建任务
└── README.md                   # 本文档
```

## 层次职责

### 1. Domains（领域层）
- **entity/**: 纯领域对象，封装业务规则和行为
- **repo/**: 仓储接口，定义数据访问规范
- **services/**: 领域服务，处理跨实体的复杂业务逻辑
  - **svorder/**: 订单服务
  - **svcallback/**: 回调服务（处理诊断回调）
  - **svdiagnosis/**: 诊断服务（发送诊断请求）
- **modules/**: 业务编排层，组合多个服务和仓储
- **apimodel/**: DTO，与外部交互的数据传输对象

### 2. Infra（基础设施层）
- **persistence/**: 实现 repo 接口，操作数据库
- **mq/**: 消息队列客户端封装

### 3. Consumer（消费者层）
- **callback_consumer.go**: 从 Lmstfy 队列消费回调消息，调用 CallbackService 处理

### 4. Server（服务器层）
- **handlers/**: HTTP 请求处理，调用 modules
- **routers/**: 路由注册
- **middlewares/**: 中间件（CORS, Logger, Error）

### 5. Pkg（通用包）
- **errorx/**: 业务错误定义
- **ginx/**: Gin 扩展（统一响应格式）
- **logger/**: 日志接口

## 数据流转

### HTTP 请求流（同步）
```
HTTP Request
  ↓
handlers/order/create.go (解析 DTO)
  ↓
modules/mdorder/order_module.go (业务编排)
  ↓
services/svorder/order_service.go (领域逻辑)
  ↓
services/svdiagnosis/diagnosis_service.go (发送诊断请求到 Lmstfy)
  ↓
repo/rporder/order_repo.go (接口)
  ↓
infra/persistence/mysql/order_repo_impl.go (实现)
  ↓
MySQL (订单数据持久化)
```

### 消息消费流（异步）
```
Lmstfy Queue (order_diagnose_callback)
  ↓
consumer/callback_consumer.go (消费消息)
  ↓
services/svcallback/callback_service.go (处理回调)
  ↓
repo/rporder/order_repo.go (更新订单状态)
  ↓
infra/persistence/redis/pubsub_client.go (发布状态变更)
  ↓
MySQL (订单状态更新) + Redis (状态通知)
```

## 命名规范

| 前缀 | 含义 | 示例 |
|------|------|------|
| `et` | Entity（实体） | `etorder.Order` |
| `md` | Module（模块） | `mdorder.OrderModule` |
| `rp` | Repository（仓储） | `rporder.OrderRepository` |
| `sv` | Service（服务） | `svorder.OrderService` |

## 快速开始

### 构建
```bash
make build
```

### 运行（单进程模式）
```bash
# 单命令启动 HTTP Server + Consumer
make run

# 或直接运行二进制文件
./bin/apiserver
```

**启动日志示例**：
```
[DPMAIN] 2024/12/28 10:00:00 [INFO] Starting callback consumer...
[DPMAIN] 2024/12/28 10:00:00 [INFO] Starting HTTP server on :8080
[DPMAIN] 2024/12/28 10:00:00 [INFO] Callback consumer started queue=order_diagnose_callback
```

### 优雅停机
```bash
# 发送 SIGINT (Ctrl+C) 或 SIGTERM
kill -TERM <pid>
```

**停机日志示例**：
```
[DPMAIN] 2024/12/28 10:05:00 [INFO] Received shutdown signal, gracefully shutting down...
[DPMAIN] 2024/12/28 10:05:00 [INFO] Stopping consumer...
[DPMAIN] 2024/12/28 10:05:01 [INFO] Stopping HTTP server...
[DPMAIN] 2024/12/28 10:05:01 [INFO] HTTP server stopped gracefully
[DPMAIN] 2024/12/28 10:05:01 [INFO] All services stopped gracefully
```

### 测试 API
```bash
# 健康检查
curl http://localhost:8080/health

# 架构说明
curl http://localhost:8080/architecture

# 创建订单（会触发诊断请求）
curl -X POST http://localhost:8080/api/v1/orders

# 查询订单
curl -X GET http://localhost:8080/api/v1/orders/123
```

## 当前状态

✅ **架构完成**
- ✅ DDD 架构目录结构完整
- ✅ 单进程多 Goroutine 架构（HTTP + Consumer）
- ✅ Wire 依赖注入完成
- ✅ 优雅停机逻辑实现
- ✅ CallbackConsumer 集成

⏳ **待完成**
- 具体业务逻辑实现
- 单元测试
- 集成测试

## 与 common 模块的关系

```
dpmain/domains/entity/etorder/  (领域对象 - 纯业务逻辑)
          ↓ 转换
common/entity/order.go          (GORM 模型 - 数据库映射)
```

- `dpmain/domains/entity`: 纯领域对象，不依赖任何框架
- `common/entity`: GORM 模型，用于数据库操作
- 在 `infra/persistence` 层进行转换
