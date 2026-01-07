# OIP Backend 测试指南

## 目录
- [环境准备](#环境准备)
- [服务启动](#服务启动)
- [测试用例](#测试用例)
- [验证方法](#验证方法)
- [常见问题](#常见问题)

---

## 环境准备

### 1. 检查 Docker 服务

确保以下 Docker 容器正在运行：

```bash
docker-compose ps
```

**预期输出：**
```
NAME         IMAGE                   STATUS
oip_mysql    mysql:8.0              Up
oip_redis    redis:7-alpine         Up
oip_lmstfy   bitleak/lmstfy:latest  Up
```

如果未运行，执行：
```bash
cd /Users/cooperswang/GolandProjects/awesomeProject/oip_backend
docker-compose up -d
```

### 2. 验证服务连接

```bash
# 测试 MySQL
docker exec -i oip_mysql mysql -uroot -ppassword -e "SELECT 1;" 2>&1 | grep -v Warning

# 测试 Redis
docker exec -i oip_redis redis-cli ping

# 测试 lmstfy
curl -s http://localhost:7777/ping || echo "lmstfy 未运行"
```

### 3. 检查数据库表

```bash
docker exec -i oip_mysql mysql -uroot -ppassword oip -e "SHOW TABLES;" 2>&1 | grep -v Warning
```

**预期输出应包含：**
- accounts
- orders

---

## 服务启动

### 方式一：使用 Makefile（推荐）

#### 1. 启动 API 服务

```bash
cd /Users/cooperswang/GolandProjects/awesomeProject/oip_backend/dpmain

# 终端 1：启动 API 服务
make run

# 或后台运行
make build
./bin/dpmain-apiserver > /tmp/dpmain-server.log 2>&1 &
```

**验证启动成功：**
```bash
curl http://localhost:8080/health
```

**预期输出：**
```json
{
  "message": "Service is running",
  "service": "dpmain",
  "status": "ok"
}
```

#### 2. 启动消费者服务

```bash
# 终端 2：启动消费者
make run-consumer

# 或后台运行
make build-consumer
./bin/dpmain-consumer > /tmp/dpmain-consumer.log 2>&1 &
```

**验证启动成功：**
```bash
tail -f /tmp/dpmain-consumer.log
```

**预期输出：**
```
Starting diagnosis consumer...
Connected to lmstfy: http://localhost:7777
Connected to Redis: localhost:6379
Consuming from queue: order_diagnose
```

### 方式二：直接运行二进制

```bash
# 编译所有服务
make build-all

# 启动 API 服务（终端 1）
./bin/dpmain-apiserver

# 启动消费者服务（终端 2）
./bin/dpmain-consumer
```

---

## 测试用例

### 测试 1: 创建账户

#### 1.1 创建测试数据文件

```bash
cat > /tmp/test_account.json <<'EOF'
{
  "name": "Test Store",
  "email": "test@example.com"
}
EOF
```

#### 1.2 发送请求

```bash
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d @/tmp/test_account.json | jq .
```

#### 1.3 预期结果

```json
{
  "code": 200,
  "data": {
    "id": 6265158001000,
    "name": "Test Store",
    "email": "test@example.com",
    "created_at": "2025-12-26T11:00:00+08:00"
  }
}
```

#### 1.4 验证数据库

```bash
docker exec -i oip_mysql mysql -uroot -ppassword oip \
  -e "SELECT * FROM accounts;" 2>&1 | grep -v Warning
```

---

### 测试 2: 查询账户

#### 2.1 使用上一步创建的账户ID

```bash
# 替换为实际的账户ID
ACCOUNT_ID=6265158001000

curl http://localhost:8080/api/v1/accounts/$ACCOUNT_ID | jq .
```

#### 2.2 预期结果

```json
{
  "code": 200,
  "data": {
    "id": 6265158001000,
    "name": "Test Store",
    "email": "test@example.com",
    "created_at": "2025-12-26T11:00:00+08:00"
  }
}
```

---

### 测试 3: 创建订单（不等待诊断）

#### 3.1 创建测试数据文件

```bash
cat > /tmp/test_order.json <<'EOF'
{
  "account_id": 6265158001000,
  "merchant_order_no": "TEST-ORDER-002",
  "shipment": {
    "ship_from": {
      "contact_name": "John's Store",
      "street1": "230 W 200 S",
      "city": "Salt Lake City",
      "state": "UT",
      "postal_code": "84101",
      "country": "US",
      "phone": "+1-801-555-0100",
      "email": "john@store.com"
    },
    "ship_to": {
      "contact_name": "Jane Doe",
      "street1": "123 Main St",
      "city": "Seattle",
      "state": "WA",
      "postal_code": "98101",
      "country": "US",
      "phone": "+1-206-555-0200",
      "email": "jane@example.com"
    },
    "parcels": [
      {
        "weight": {"value": 1.5, "unit": "kg"},
        "dimension": {"width": 20, "height": 15, "depth": 10, "unit": "cm"},
        "items": [
          {
            "description": "Wireless Mouse",
            "quantity": 2,
            "price": {"amount": 19.99, "currency": "USD"},
            "sku": "MOUSE-WL-001"
          }
        ]
      }
    ]
  }
}
EOF
```

**注意：** 请将 `account_id` 替换为实际创建的账户ID。

#### 3.2 发送请求（不等待诊断）

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d @/tmp/test_order.json | jq .
```

#### 3.3 预期结果（3001 Processing）

```json
{
  "code": 3001,
  "message": "Order is being diagnosed, please poll for results",
  "data": {
    "order_id": "cb7ecba8-6fc4-4c70-bf2d-14d3caa31a32"
  },
  "poll_url": "/api/v1/orders/cb7ecba8-6fc4-4c70-bf2d-14d3caa31a32"
}
```

---

### 测试 4: 创建订单（Smart Wait 模式）

#### 4.1 修改订单号（避免重复）

```bash
cat > /tmp/test_order_wait.json <<'EOF'
{
  "account_id": 6265158001000,
  "merchant_order_no": "TEST-ORDER-002",
  "shipment": {
    "ship_from": {
      "contact_name": "John's Store",
      "street1": "230 W 200 S",
      "city": "Salt Lake City",
      "state": "UT",
      "postal_code": "84101",
      "country": "US",
      "phone": "+1-801-555-0100",
      "email": "john@store.com"
    },
    "ship_to": {
      "contact_name": "Jane Doe",
      "street1": "123 Main St",
      "city": "Seattle",
      "state": "WA",
      "postal_code": "98101",
      "country": "US",
      "phone": "+1-206-555-0200",
      "email": "jane@example.com"
    },
    "parcels": [
      {
        "weight": {"value": 1.5, "unit": "kg"},
        "dimension": {"width": 20, "height": 15, "depth": 10, "unit": "cm"},
        "items": [
          {
            "description": "Wireless Mouse",
            "quantity": 2,
            "price": {"amount": 19.99, "currency": "USD"},
            "sku": "MOUSE-WL-001"
          }
        ]
      }
    ]
  }
}
EOF
```

#### 4.2 发送请求（等待10秒）

```bash
curl -X POST "http://localhost:8080/api/v1/orders?wait=10" \
  -H "Content-Type: application/json" \
  -d @/tmp/test_order_wait.json | jq .
```

#### 4.3 预期结果（200 OK - 诊断完成）

**如果消费者服务正常运行，应该在2-3秒内返回：**

```json
{
  "code": 200,
  "data": {
    "id": "2a5fe3fc-d5c3-46b0-91fd-a7960c5091a2",
    "account_id": 6265158001000,
    "merchant_order_no": "TEST-ORDER-002",
    "status": "DIAGNOSED",
    "diagnosis": {
      "items": [
        {
          "type": "shipping",
          "status": "SUCCESS"
        },
        {
          "type": "anomaly",
          "status": "SUCCESS"
        }
      ]
    },
    "created_at": "2025-12-26T11:37:46+08:00",
    "updated_at": "2025-12-26T11:37:48+08:00"
  }
}
```

#### 4.4 同时观察消费者日志

在另一个终端执行：
```bash
tail -f /tmp/dpmain-consumer.log
```

**预期看到：**
```
[INFO] Received message: job_id=01KDCBQQWTXBJ2ZBYP9W000000
[INFO] Processing order: 2a5fe3fc-d5c3-46b0-91fd-a7960c5091a2
[SUCCESS] Published diagnosis result for order 2a5fe3fc-d5c3-46b0-91fd-a7960c5091a2
[INFO] Message acknowledged: job_id=01KDCBQQWTXBJ2ZBYP9W000000
```

---

### 测试 5: 查询订单详情

#### 5.1 使用上一步创建的订单ID

```bash
# 替换为实际的订单ID
ORDER_ID="2a5fe3fc-d5c3-46b0-91fd-a7960c5091a2"

curl http://localhost:8080/api/v1/orders/$ORDER_ID | jq .
```

#### 5.2 预期结果

```json
{
  "code": 200,
  "data": {
    "id": "2a5fe3fc-d5c3-46b0-91fd-a7960c5091a2",
    "account_id": 6265158001000,
    "merchant_order_no": "TEST-ORDER-002",
    "status": "DIAGNOSED",
    "diagnosis": {
      "items": [...]
    },
    "created_at": "2025-12-26T11:37:46+08:00",
    "updated_at": "2025-12-26T11:37:48+08:00"
  }
}
```

---

## 验证方法

### 1. 检查服务进程

```bash
# 检查 API 服务
ps aux | grep dpmain-apiserver | grep -v grep

# 检查消费者服务
ps aux | grep dpmain-consumer | grep -v grep

# 检查端口占用
lsof -i :8080  # API 服务
```

### 2. 查看实时日志

```bash
# API 服务日志
tail -f /tmp/dpmain-server.log

# 消费者日志
tail -f /tmp/dpmain-consumer.log
```

### 3. 验证数据库数据

```bash
# 查看所有账户
docker exec -i oip_mysql mysql -uroot -ppassword oip \
  -e "SELECT * FROM accounts;" 2>&1 | grep -v Warning

# 查看所有订单
docker exec -i oip_mysql mysql -uroot -ppassword oip \
  -e "SELECT id, account_id, merchant_order_no, status FROM orders;" 2>&1 | grep -v Warning
```

### 4. 验证消息队列

```bash
# 查看 lmstfy 状态（需要 token）
curl -H "X-Token: 01KDCBF5BG0THBC24F1V53XPR1" \
  "http://localhost:7777/api/oip/order_diagnose?timeout=1&ttr=30"
```

### 5. 验证 Redis 连接

```bash
docker exec -i oip_redis redis-cli ping
```

---

## 常见问题

### 问题 1: 端口被占用

**错误信息：**
```
listen tcp :8080: bind: address already in use
```

**解决方法：**
```bash
# 查找占用进程
lsof -i :8080

# 杀掉进程
kill -9 <PID>
```

---

### 问题 2: 数据库表不存在

**错误信息：**
```
Error 1146: Table 'oip.accounts' doesn't exist
```

**解决方法：**
```bash
# 重新导入表结构
docker exec -i oip_mysql mysql -uroot -ppassword oip < sql/schema.sql
```

---

### 问题 3: lmstfy 认证失败

**错误信息：**
```
lmstfy consume failed: status=401
```

**解决方法：**
```bash
# 创建新的 token
curl -X POST "http://localhost:7778/token/oip?description=test-token"

# 更新代码中的 token（wire.go 和 consumer/main.go）
# 重新编译并启动服务
```

---

### 问题 4: 消费者无法启动

**检查步骤：**

1. 验证 lmstfy 服务运行：
```bash
docker ps | grep lmstfy
curl http://localhost:7777/ping
```

2. 验证 Redis 服务运行：
```bash
docker ps | grep redis
docker exec -i oip_redis redis-cli ping
```

3. 检查消费者日志：
```bash
tail -20 /tmp/dpmain-consumer.log
```

---

### 问题 5: Smart Wait 超时（返回 3001）

**可能原因：**
1. 消费者服务未启动
2. 消费者处理消息失败
3. Redis Pub/Sub 连接问题

**排查步骤：**

1. 检查消费者是否运行：
```bash
ps aux | grep dpmain-consumer
```

2. 查看消费者日志：
```bash
tail -30 /tmp/dpmain-consumer.log
```

3. 手动测试 Redis Pub/Sub：
```bash
# 终端 1：订阅
docker exec -i oip_redis redis-cli SUBSCRIBE test_channel

# 终端 2：发布
docker exec -i oip_redis redis-cli PUBLISH test_channel "hello"
```

---

## 快速测试脚本

创建一个一键测试脚本：

```bash
cat > /tmp/quick_test.sh <<'EOF'
#!/bin/bash

echo "=== OIP Backend 快速测试 ==="
echo ""

# 1. 测试健康检查
echo "1. 测试健康检查..."
curl -s http://localhost:8080/health | jq .
echo ""

# 2. 创建账户
echo "2. 创建测试账户..."
ACCOUNT_RESPONSE=$(curl -s -X POST http://localhost:8080/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{"name":"Quick Test","email":"quicktest@example.com"}')
echo $ACCOUNT_RESPONSE | jq .
ACCOUNT_ID=$(echo $ACCOUNT_RESPONSE | jq -r '.data.id')
echo "账户ID: $ACCOUNT_ID"
echo ""

# 3. 创建订单（Smart Wait）
echo "3. 创建订单（Smart Wait 10秒）..."
ORDER_RESPONSE=$(curl -s -X POST "http://localhost:8080/api/v1/orders?wait=10" \
  -H "Content-Type: application/json" \
  -d "{
    \"account_id\": $ACCOUNT_ID,
    \"merchant_order_no\": \"QUICK-TEST-$(date +%s)\",
    \"shipment\": {
      \"ship_from\": {
        \"contact_name\": \"Test Store\",
        \"street1\": \"123 Test St\",
        \"city\": \"Test City\",
        \"state\": \"CA\",
        \"postal_code\": \"12345\",
        \"country\": \"US\"
      },
      \"ship_to\": {
        \"contact_name\": \"Test Customer\",
        \"street1\": \"456 Test Ave\",
        \"city\": \"Test Town\",
        \"state\": \"NY\",
        \"postal_code\": \"67890\",
        \"country\": \"US\"
      },
      \"parcels\": [{
        \"weight\": {\"value\": 1.0, \"unit\": \"kg\"},
        \"items\": [{
          \"description\": \"Test Item\",
          \"quantity\": 1,
          \"price\": {\"amount\": 10.00, \"currency\": \"USD\"}
        }]
      }]
    }
  }")
echo $ORDER_RESPONSE | jq .
echo ""

echo "=== 测试完成 ==="
EOF

chmod +x /tmp/quick_test.sh
```

**运行快速测试：**
```bash
/tmp/quick_test.sh
```

---

## 测试清单

使用此清单确保所有功能正常：

- [ ] Docker 服务运行正常
- [ ] API 服务启动成功
- [ ] 消费者服务启动成功
- [ ] 健康检查接口返回 200
- [ ] 创建账户成功
- [ ] 查询账户成功
- [ ] 创建订单（不等待）返回 3001
- [ ] 创建订单（Smart Wait）返回 200 + 诊断结果
- [ ] 查询订单详情成功
- [ ] 消费者日志显示消息处理成功
- [ ] 数据库中账户和订单数据正确

---

## 联系信息

如有问题，请检查：
1. `/tmp/dpmain-server.log` - API 服务日志
2. `/tmp/dpmain-consumer.log` - 消费者日志
3. `docker logs oip_lmstfy` - lmstfy 日志
4. `docker logs oip_redis` - Redis 日志
5. `docker logs oip_mysql` - MySQL 日志

---

**测试指南版本：** v1.0
**最后更新：** 2025-12-26
