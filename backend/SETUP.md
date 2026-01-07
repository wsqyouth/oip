# OIP Backend - 架构验证指南

## 环境要求

- Go 1.21+
- Make

## 重要：Go 环境配置

**注意**：在运行构建前，请确保 Go 环境变量配置正确。

### 检查当前配置

```bash
go env GOPATH GOMODCACHE
```

### 推荐配置（添加到 ~/.zshrc 或 ~/.bashrc）

```bash
export GOPATH=$HOME/go
export GOMODCACHE=$GOPATH/pkg/mod
export PATH=$PATH:$GOPATH/bin
```

配置后重新加载：
```bash
source ~/.zshrc  # 或 source ~/.bashrc
```

## 快速验证

### 方式 1：使用验证脚本（推荐）

```bash
cd /Users/cooperswang/GolandProjects/awesomeProject/oip_backend
./scripts/verify.sh
```

### 方式 2：使用 Makefile

```bash
cd /Users/cooperswang/GolandProjects/awesomeProject/oip_backend

# 构建所有模块
make build

# 测试
make test
```

## 项目结构

```
oip_backend/
├── go.work              # Go Workspace 配置
├── Makefile             # 根 Makefile
├── docker-compose.yml   # 基础设施（MySQL, Redis, Lmstfy）
│
├── common/              # 共享内核模块
│   ├── go.mod
│   ├── entity/          # GORM 数据模型（Order, Account）
│   ├── model/           # 诊断结果结构体（DiagnosisResult）
│   └── dao/             # 数据访问层（OrderDAO, AccountDAO）
│
├── dpmain/              # 同步 API 服务
│   ├── go.mod
│   ├── cmd/apiserver/   # 程序入口（main.go）
│   ├── internal/
│   │   ├── api/         # HTTP Handlers（OrderHandler, AccountHandler）
│   │   ├── service/     # 业务服务层（OrderService）
│   │   └── middleware/  # 中间件（CORS）
│   ├── pkg/
│   │   ├── config/      # 配置管理
│   │   ├── redis/       # Redis 客户端封装
│   │   └── logger/      # 日志（预留）
│   └── Makefile
│
└── dpsync/              # 异步 Worker 服务
    ├── go.mod
    ├── cmd/worker/      # 程序入口（main.go）
    ├── internal/
    │   ├── worker/      # Worker 核心逻辑
    │   └── handlers/    # 业务处理器
    │       ├── composite_handler.go   # 组合处理器
    │       ├── shipping_calculator.go # 物流费率计算（Mock）
    │       └── anomaly_checker.go     # 异常检测
    ├── pkg/
    │   ├── config/      # 配置管理
    │   ├── lmstfy/      # Lmstfy 客户端封装
    │   └── logger/      # 日志（预留）
    ├── config/
    │   └── worker.yaml  # Worker 配置示例
    └── Makefile
```

## 架构特点

### 1. Monorepo + Go Workspace
- 使用 `go.work` 管理多个模块
- `common` 模块通过 `replace` 指令被 `dpmain` 和 `dpsync` 引用

### 2. 模块职责划分
- **common**: 共享的数据模型和 DAO 层（无业务逻辑）
- **dpmain**: 同步 HTTP API 服务（Smart Wait 机制）
- **dpsync**: 异步 Worker 服务（诊断任务处理）

### 3. 预留的扩展点
所有业务逻辑都标记为 `TODO`，包括：
- Order/Account 的 CRUD 实现
- Smart Wait（Redis Pub/Sub）
- CompositeHandler 诊断逻辑
- Mock 费率计算和异常检测

## 验证清单

运行 `./scripts/verify.sh` 后，应该看到：

- [x] Go 版本检查通过（1.21+）
- [x] common 模块 `go mod tidy` 成功
- [x] dpmain 模块构建成功（生成 `dpmain/bin/dpmain-apiserver`）
- [x] dpsync 模块构建成功（生成 `dpsync/bin/dpsync-worker`）
- [x] Go Workspace 同步成功

## 下一步

架构验证通过后，可以开始：

1. 实现 `common/dao` 的业务逻辑
2. 实现 `dpmain/internal/api` 的 HTTP 接口
3. 实现 `dpsync/internal/handlers` 的诊断逻辑
4. 参考 `PRD_v3.0.md` 文档完成 MVP 开发

## 常见问题

### Q: go mod tidy 报错 "permission denied"
A: 检查 `GOPATH` 和 `GOMODCACHE` 环境变量是否指向有权限的目录。

### Q: 如何单独构建某个模块？
A:
```bash
cd dpmain && make build
cd dpsync && make build
```

### Q: 如何启动服务？
A:
```bash
# 启动 API 服务（端口 8080）
make run-dpmain

# 启动 Worker 服务
make run-dpsync

# 启动基础设施（MySQL, Redis, Lmstfy）
make docker-compose-up
```
