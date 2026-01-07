# E2E Test - DPSYNC 端到端测试

## 功能

端到端测试脚本，用于验证从 lmstfy 队列到数据库的完整链路。

**测试流程：**
1. 发送测试消息到 lmstfy 队列
2. Worker 消费消息
3. 执行诊断逻辑
4. 更新数据库
5. 发送 Redis 通知
6. 验证结果

## 前置条件

1. **启动 lmstfy 服务**

```bash
docker run -d -p 7777:7777 bitleak/lmstfy
```

2. **启动 MySQL 服务**

```bash
# 使用 docker-compose 启动
cd /Users/cooperswang/GolandProjects/awesomeProject/oip_backend
docker-compose up -d mysql

# 或手动启动 MySQL 容器
docker run -d -p 3306:3306 \
  -e MYSQL_ROOT_PASSWORD=password \
  -e MYSQL_DATABASE=oip \
  mysql:8.0
```

3. **启动 Redis 服务**

```bash
docker-compose up -d redis

# 或手动启动 Redis 容器
docker run -d -p 6379:6379 redis:7
```

4. **启动 DPSYNC Worker**

在一个终端窗口中启动 Worker：

```bash
cd /Users/cooperswang/GolandProjects/awesomeProject/oip_backend/dpsync
go run cmd/worker/main.go
```

## 使用方法

### 运行端到端测试

```bash
cd /Users/cooperswang/GolandProjects/awesomeProject/oip_backend/dpsync
./tools/e2etest/run_e2e_test.sh
```

### 自定义配置

可以通过环境变量自定义配置：

```bash
LMSTFY_HOST=http://localhost:7777 \
MYSQL_DSN=root:password@tcp(127.0.0.1:3306)/oip \
REDIS_ADDR=localhost:6379 \
./tools/e2etest/run_e2e_test.sh
```

## 输出示例

```
========================================
  DPSYNC 端到端测试
========================================
📝 测试配置：
  - lmstfy: http://localhost:7777
  - Queue: oip_order_diagnose
  - OrderID: e2e_test_1703001234
  - AccountID: 999

🔍 [Step 1] 检查依赖服务...
  - lmstfy: ✅ Running

📦 [Step 2] 构造测试消息...
消息内容：
{
  "payload": {
    "data": {
      "request_id": "e2e-test-1703001234",
      "org_id": "org-test",
      "action_type": "order_diagnose",
      "id": "diag-e2e-test",
      "data": {
        "order_id": "e2e_test_1703001234",
        "account_id": 999
      }
    }
  }
}

📨 [Step 3] 发送消息到 lmstfy...
✅ 消息发送成功
  - Job ID: 01HJKM5N6QXYZ

⏳ [Step 4] 等待 Worker 处理消息（最多 30 秒）...
  请确保 Worker 正在运行：go run cmd/worker/main.go

.............................. Done

🔍 [Step 5] 验证数据库结果（可选）...
检查订单诊断结果...
+-------------------+-----------+---------------------------+
| id                | status    | types                     |
+-------------------+-----------+---------------------------+
| e2e_test_17030... | DIAGNOSED | ["shipping", "anomaly"]   |
+-------------------+-----------+---------------------------+

🔍 [Step 6] 验证 Redis 通知（可选）...
订阅 Redis 频道 'order_diagnosis_complete' 查看通知：
  redis-cli SUBSCRIBE order_diagnosis_complete

========================================
  测试汇总
========================================
✅ 测试消息已发送到 lmstfy
⏳ Worker 应该在 30 秒内处理完消息

手动验证步骤：
1. 检查 Worker 日志，确认消息被处理
2. 查询数据库：SELECT * FROM orders WHERE id = 'e2e_test_1703001234';
3. 订阅 Redis：redis-cli SUBSCRIBE order_diagnosis_complete

如果以上步骤都成功，说明端到端测试通过！🎉
========================================
```

## 手动验证步骤

### 1. 查看 Worker 日志

在运行 Worker 的终端中，应该看到类似输出：

```
=== DiagnoseHandler Process ===
{
  "handler": "DiagnoseHandler",
  "action": "order_diagnose",
  "order_id": "e2e_test_1703001234",
  "account_id": 999,
  "phase": "Phase 3: Full diagnosis with DB & Redis"
}
Diagnosis completed successfully:
  - Items: 2
  [1] Type=shipping, Status=SUCCESS
  [2] Type=anomaly, Status=SUCCESS
  - DB updated: YES
  - Redis notified: YES
==============================
```

### 2. 查询数据库

```bash
mysql -h 127.0.0.1 -u root -ppassword oip
```

```sql
SELECT
  id,
  status,
  JSON_PRETTY(diagnose_result) as result,
  updated_at
FROM orders
WHERE id LIKE 'e2e_test_%'
ORDER BY created_at DESC
LIMIT 5;
```

### 3. 监听 Redis 通知

在另一个终端窗口中：

```bash
redis-cli SUBSCRIBE order_diagnosis_complete
```

运行测试后，应该收到类似通知：

```
1) "message"
2) "order_diagnosis_complete"
3) "{\"order_id\":\"e2e_test_1703001234\",\"account_id\":999,\"status\":\"DIAGNOSED\",\"timestamp\":1703001234}"
```

## 故障排查

### 问题：lmstfy 连接失败

**症状：**
```
❌ lmstfy: Not running
```

**解决方案：**
```bash
docker run -d -p 7777:7777 bitleak/lmstfy
```

### 问题：Worker 未处理消息

**症状：** 30 秒后数据库中没有记录

**排查步骤：**
1. 确认 Worker 正在运行
2. 检查 Worker 日志是否有错误
3. 检查队列名称是否匹配（`oip_order_diagnose`）
4. 使用 FastTest 工具验证业务逻辑是否正常

### 问题：数据库连接失败

**症状：** Worker 启动时报错 "Failed to create order DAO"

**解决方案：**
1. 确认 MySQL 服务正在运行
2. 检查 `config/worker.yaml` 中的 DSN 配置
3. 验证数据库 `oip` 是否已创建
4. 验证 `orders` 表是否已创建

### 问题：Redis 连接失败

**症状：** Worker 启动时报错 "Failed to create redis pubsub"

**解决方案：**
1. 确认 Redis 服务正在运行
2. 检查 `config/worker.yaml` 中的 Redis 配置

## 性能基准

正常情况下：
- 消息发送到 lmstfy：< 10ms
- Worker 拉取消息：< 100ms
- 诊断逻辑执行：10-20ms
- 数据库更新：20-50ms
- Redis 通知：< 5ms
- **总耗时**：约 50-200ms

如果处理时间明显超过 200ms，请检查：
- 数据库连接池配置
- 网络延迟
- Worker 并发配置
