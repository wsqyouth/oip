# OIP Backend - 跨境订单智能诊断平台

## 项目结构（Monorepo）

```
oip_backend/
├── go.work              # Go Workspace 配置
├── common/              # 共享内核模块
│   ├── entity/          # GORM 数据模型
│   ├── model/           # 诊断结果结构体
│   └── dao/             # 数据访问层
├── dpmain/              # 同步 API 服务
│   ├── cmd/             # 程序入口
│   ├── internal/        # 内部业务逻辑
│   └── pkg/             # 可复用包
└── dpsync/              # 异步 Worker 服务
    ├── cmd/             # 程序入口
    ├── internal/        # 内部业务逻辑
    └── pkg/             # 可复用包
```

## 快速开始

### 构建所有模块
```bash
make build
```

### 运行服务
```bash
# 启动 dpmain（API 服务）
make run-dpmain

# 启动 dpsync（Worker 服务）
make run-dpsync
```

### 测试
```bash
make test
```

## 开发说明

本项目使用 Go Workspace 管理 Monorepo：
- `common`: 共享的数据模型和 DAO 层
- `dpmain`: 同步 HTTP API 服务
- `dpsync`: 异步诊断 Worker 服务

每个模块都是独立的 Go Module，可以单独构建和测试。
