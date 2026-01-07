# FastTest - DPSYNC Worker 快速测试工具

## 功能

FastTest 是一个快速测试工具，用于验证 DPSYNC Worker 的业务逻辑，无需启动完整的 Worker 和 lmstfy 队列。

**核心特性：**
- 直接调用业务逻辑，跳过消息队列
- 支持两种模式：
  - **Skip-DB 模式**：仅测试 CompositeHandler 业务逻辑
  - **完整模式**：测试完整流程（诊断 + 数据库 + Redis）
- 从 JSON 文件加载测试用例
- 输出详细的测试结果和性能指标

## 使用方法

### 1. Skip-DB 模式（仅测试业务逻辑）

适用于快速验证诊断逻辑，无需数据库和 Redis 环境：

```bash
cd /Users/cooperswang/GolandProjects/awesomeProject/oip_backend/dpsync

go run tools/fasttest/worker_fast_test.go --skip-db
```

### 2. 完整模式（测试完整流程）

需要启动 MySQL 和 Redis 服务：

```bash
# 确保 MySQL 和 Redis 已启动
docker-compose up -d mysql redis

# 运行 FastTest
go run tools/fasttest/worker_fast_test.go
```

### 3. 指定自定义配置和测试用例

```bash
go run tools/fasttest/worker_fast_test.go \
  --config ./config/worker.yaml \
  --testcase ./internal/domains/handlers/order/diagnose/testcase/diagnose.json
```

## 参数说明

| 参数 | 默认值 | 说明 |
|------|--------|------|
| `--config` | `./config/worker.yaml` | Worker 配置文件路径 |
| `--testcase` | `./internal/domains/handlers/order/diagnose/testcase/diagnose.json` | 测试用例文件路径 |
| `--skip-db` | `false` | 是否跳过数据库和 Redis 操作 |

## 测试用例格式

测试用例文件为 JSON 数组，每个元素包含：

```json
[
  {
    "order_id": "ord_550e8400e29b41d4",
    "account_id": 1
  },
  {
    "order_id": "ord_661f9511f3ac52e5",
    "account_id": 2
  }
]
```

## 输出示例

### Skip-DB 模式

```
========================================
  FastTest - DPSYNC Worker 快速测试工具
========================================
✅ Config loaded: dpsync-worker
✅ Loaded 2 test cases from ./internal/domains/handlers/order/diagnose/testcase/diagnose.json
⚠️  Skip-DB mode: Database and Redis operations disabled

========================================
  Running Test Cases
========================================

[Test 1/2] OrderID=ord_550e8400e29b41d4, AccountID=1
----------------------------------------
  Diagnosis Items: 2
    - Type=shipping, Status=SUCCESS
    - Type=anomaly, Status=SUCCESS
✅ PASSED
⏱️  Duration: 12ms

[Test 2/2] OrderID=ord_661f9511f3ac52e5, AccountID=2
----------------------------------------
  Diagnosis Items: 2
    - Type=shipping, Status=SUCCESS
    - Type=anomaly, Status=SUCCESS
✅ PASSED
⏱️  Duration: 8ms

========================================
  Test Summary
========================================
Total: 2
Passed: 2 ✅
Failed: 0 ❌
```

### 完整模式

```
========================================
  FastTest - DPSYNC Worker 快速测试工具
========================================
✅ Config loaded: dpsync-worker
✅ Loaded 2 test cases from ./internal/domains/handlers/order/diagnose/testcase/diagnose.json
✅ Database and Redis initialized

========================================
  Running Test Cases
========================================

[Test 1/2] OrderID=ord_550e8400e29b41d4, AccountID=1
----------------------------------------
  Diagnosis Items: 2
    - Type=shipping, Status=SUCCESS
    - Type=anomaly, Status=SUCCESS
  ✓ Database updated
  ✓ Redis notification sent
✅ PASSED
⏱️  Duration: 45ms

[Test 2/2] OrderID=ord_661f9511f3ac52e5, AccountID=2
----------------------------------------
  Diagnosis Items: 2
    - Type=shipping, Status=SUCCESS
    - Type=anomaly, Status=SUCCESS
  ✓ Database updated
  ✓ Redis notification sent
✅ PASSED
⏱️  Duration: 38ms

========================================
  Test Summary
========================================
Total: 2
Passed: 2 ✅
Failed: 0 ❌
```

## 注意事项

1. **Skip-DB 模式**：
   - 不需要数据库和 Redis 环境
   - 仅测试 ShippingCalculator 和 AnomalyChecker 逻辑
   - 适用于快速验证业务规则

2. **完整模式**：
   - 需要 MySQL 和 Redis 服务运行
   - 会实际更新数据库中的订单诊断结果
   - 会发送 Redis 通知到 `order_diagnosis_complete` 频道
   - 适用于端到端测试验证

3. **性能基准**：
   - Skip-DB 模式：单个测试用例通常在 5-15ms
   - 完整模式：单个测试用例通常在 30-60ms（含数据库 I/O）

## 扩展测试用例

在 `internal/domains/handlers/order/diagnose/testcase/diagnose.json` 中添加更多测试用例：

```json
[
  {
    "order_id": "ord_550e8400e29b41d4",
    "account_id": 1
  },
  {
    "order_id": "remote_order_12345",
    "account_id": 2
  },
  {
    "order_id": "short",
    "account_id": 3
  }
]
```

不同的 order_id 会触发不同的诊断规则（参见 `AnomalyChecker` 实现）。
