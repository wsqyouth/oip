# OIP Backend E2E 测试指南

## 1. 概述

### 1.1 文档目的
本文档提供 OIP Backend 完整的端到端（E2E）测试指南，帮助开发者验证从订单创建到诊断完成的整个业务链路。

### 1.2 测试范围
本 E2E 测试覆盖以下完整链路：

```
客户端
  ↓ POST /api/v1/orders
[dpmain] API Server
  ↓ 1. 接收订单请求
  ↓ 2. 保存订单到 MySQL
  ↓ 3. 推送任务到 Lmstfy (order_diagnose)
  ↓ 4. 订阅 Redis channel 等待结果
  ↓
Lmstfy Queue (order_diagnose)
  ↓
[dpsync] Worker
  ↓ 5. 消费诊断任务
  ↓ 6. 执行诊断逻辑 (shipping, anomaly)
  ↓ 7. 推送回调到 Lmstfy (order_diagnose_callback)
  ↓
Lmstfy Queue (order_diagnose_callback)
  ↓
[dpmain] Callback Consumer
  ↓ 8. 消费回调消息
  ↓ 9. 更新订单状态到 MySQL
  ↓ 10. 发布结果到 Redis Pub/Sub
  ↓
[dpmain] API Server
  ↓ 接收 Redis 通知
  ↓ 返回诊断结果
  ↓
客户端 (收到完整诊断结果)
```

### 1.3 前置条件

**必需环境：**
- Go 1.21+
- Docker 和 Docker Compose
- curl 命令行工具

**可选工具：**
- jq（用于 JSON 格式化和解析）
  ```bash
  # macOS
  brew install jq
  
  # Linux
  apt-get install jq
  ```

## 2. 测试架构说明

### 2.1 服务组件

| 组件 | 说明 | 端口 |
|------|------|------|
| **dpmain** | API 服务 + Callback Consumer（单进程） | 8080 |
| **dpsync** | 诊断 Worker 服务 | - |
| **MySQL** | 数据持久化 | 3306 |
| **Redis** | Pub/Sub 通知 | 6379 |
| **Lmstfy** | 消息队列 | 7777 |

### 2.2 端口映射

```
localhost:8080  → dpmain API Server
localhost:3306  → MySQL (oip 数据库)
localhost:6379  → Redis
localhost:7777  → Lmstfy HTTP API
```

---

# 第一部分：手动 E2E 测试

## 3. 环境准备

### 3.1 启动依赖服务

#### Step 1: 启动 Docker Compose

```bash
cd /Users/cooperswang/Documents/wsqyouth/oip/backend
docker-compose up -d
```

**预期输出：**
```
[+] Running 3/3
 ✔ Container oip_mysql   Started
 ✔ Container oip_redis   Started
 ✔ Container oip_lmstfy  Started
```

#### Step 2: 验证容器状态

```bash
docker-compose ps
```

**预期输出：**
```
NAME         IMAGE                   STATUS
oip_lmstfy   bitleak/lmstfy:latest  Up
oip_mysql    mysql:8.0              Up (healthy)
oip_redis    redis:7-alpine         Up
```

#### Step 3: 测试服务连通性

**测试 MySQL：**
```bash
docker exec -i oip_mysql mysql -uroot -ppassword -e "SELECT 1;" 2>&1 | grep -v Warning
```

**预期输出：**
```
1
1
```

**测试 Redis：**
```bash
docker exec -i oip_redis redis-cli ping
```

**预期输出：**
```
PONG
```

**测试 Lmstfy：**
```bash
curl -s http://localhost:7777/ping
```

**预期输出：**
```
pong
```

### 3.2 构建应用服务

#### 构建 dpmain

```bash
cd /Users/cooperswang/Documents/wsqyouth/oip/backend/dpmain
go mod tidy
go build -o bin/dpmain-apiserver ./cmd/apiserver
```

**验证构建成功：**
```bash
ls -lh bin/dpmain-apiserver
```

#### 构建 dpsync

```bash
cd /Users/cooperswang/Documents/wsqyouth/oip/backend/dpsync
go mod tidy
go build -o bin/dpsync-worker ./cmd/worker
```

**验证构建成功：**
```bash
ls -lh bin/dpsync-worker
```

### 3.3 启动应用服务

#### 终端 1: 启动 dpmain

```bash
cd /Users/cooperswang/Documents/wsqyouth/oip/backend/dpmain
./bin/dpmain-d
```

**预期日志输出：**
```
[INFO] 2025-01-09 10:00:00 Starting dpmain API server...
[INFO] 2025-01-09 10:00:00 Database connected: mysql://localhost:3306/oip
[INFO] 2025-01-09 10:00:00 Redis connected: localhost:6379
[INFO] 2025-01-09 10:00:00 Lmstfy connected: http://localhost:7777
[INFO] 2025-01-09 10:00:00 Starting callback consumer...
[INFO] 2025-01-09 10:00:00 Callback consumer started, queue: order_diagnose_callback
[INFO] 2025-01-09 10:00:00 HTTP server listening on :8080
```

#### 终端 2: 启动 dpsync

```bash
cd /Users/cooperswang/Documents/wsqyouth/oip/backend/dpsync
./bin/dpsync-worker
```

**预期日志输出：**
```
[INFO] 2025-01-09 10:00:05 Starting dpsync worker...
[INFO] 2025-01-09 10:00:05 Database connected: mysql://localhost:3306/oip
[INFO] 2025-01-09 10:00:05 Lmstfy connected: http://localhost:7777
[INFO] 2025-01-09 10:00:05 Worker started, consuming queue: order_diagnose
[INFO] 2025-01-09 10:00:05 Waiting for messages...
```

#### 验证服务健康状态

**终端 3：**
```bash
curl http://localhost:8080/health
```

**预期响应：**
```json
{
  "message": "Service is running",
  "service": "dpmain",
  "status": "ok"
}
```

---

## 4. 执行 E2E 测试

### 4.1 测试 1：创建账户

#### 执行命令

```bash
curl -X POST http://localhost:8080/api/v1/accounts \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Store",
    "email": "test@example.com"
  }' | jq .
```

#### 预期响应（JSON 格式）

```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "id": 6265158001000,
    "name": "Test Store",
    "email": "test@example.com",
    "created_at": "2025-01-09T10:00:00+08:00",
    "updated_at": "2025-01-09T10:00:00+08:00"
  }
}
```

#### API 响应格式说明

**成功响应结构：**
```json
{
  "code": 200,          // 状态码
  "message": "Success", // 响应消息
  "data": { ... }       // 业务数据
}
```

**错误响应结构：**
```json
{
  "code": 400,                    // 错误码
  "message": "Validation failed", // 错误消息
  "details": [                    // 详细错误信息（可选）
    {
      "path": "email",
      "info": "email is required"
    }
  ]
}
```

#### 数据库验证（可选）

```bash
docker exec -i oip_mysql mysql -uroot -ppassword oip \
  -e "SELECT id, name, email FROM accounts ORDER BY id DESC LIMIT 1;" \
  2>&1 | grep -v Warning
```

**保存账户 ID 到环境变量：**
```bash
export ACCOUNT_ID=6388056201000  # 替换为实际返回的 ID
```

---

### 4.2 测试 2：查询账户

#### 执行命令

```bash
curl http://localhost:8080/api/v1/accounts/$ACCOUNT_ID | jq .
```

#### 预期响应

```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "id": 6265158001000,
    "name": "Test Store",
    "email": "test@example.com",
    "created_at": "2025-01-09T10:00:00+08:00",
    "updated_at": "2025-01-09T10:00:00+08:00"
  }
}
```

---

### 4.3 测试 3：创建订单（不等待诊断）

#### 准备测试数据

创建测试数据文件：

```bash
cat > /tmp/test_order_no_wait.json <<'EOF'
{
  "account_id": 6265158001000,
  "merchant_order_no": "TEST-ORDER-NO-WAIT-001",
  "shipment": {
    "ship_from": {
      "contact_name": "Test Store",
      "street1": "230 W 200 S",
      "city": "Salt Lake City",
      "state": "UT",
      "postal_code": "84101",
      "country": "US",
      "phone": "+1-801-555-0100",
      "email": "store@test.com"
    },
    "ship_to": {
      "contact_name": "John Doe",
      "street1": "123 Main St",
      "city": "Seattle",
      "state": "WA",
      "postal_code": "98101",
      "country": "US",
      "phone": "+1-206-555-0200",
      "email": "john@example.com"
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

**注意：** 替换 `account_id` 为实际的账户 ID

#### 执行命令

```bash
curl -X POST http://localhost:8080/api/v1/orders \
  -H "Content-Type: application/json" \
  -d @/tmp/test_order_no_wait.json | jq .
```

#### 预期响应（3001 Processing）

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

#### 说明：异步处理场景

- **Code 3001**：表示订单已创建，诊断任务已推送到队列，但在默认等待时间内未完成
- 这是**正常的异步处理流程**
- 客户端应通过 `poll_url` 轮询获取诊断结果

---

### 4.4 测试 4：创建订单（Smart Wait 模式）**【核心测试】**

这是最重要的测试，验证完整的诊断链路。

#### 准备测试数据

创建测试数据文件：

```bash
cat > /tmp/test_order_smart_wait.json <<'EOF'
{
  "account_id": 6388056201000,
  "merchant_order_no": "TEST-ORDER-SMART-WAIT-001",
  "shipment": {
    "ship_from": {
      "contact_name": "Test Store",
      "street1": "230 W 200 S",
      "city": "Salt Lake City",
      "state": "UT",
      "postal_code": "84101",
      "country": "US",
      "phone": "+1-801-555-0100",
      "email": "store@test.com"
    },
    "ship_to": {
      "contact_name": "Jane Smith",
      "street1": "456 Oak Ave",
      "city": "Los Angeles",
      "state": "CA",
      "postal_code": "90001",
      "country": "US",
      "phone": "+1-213-555-0300",
      "email": "jane@example.com"
    },
    "parcels": [
      {
        "weight": {"value": 2.0, "unit": "kg"},
        "dimension": {"width": 25, "height": 20, "depth": 15, "unit": "cm"},
        "items": [
          {
            "description": "Mechanical Keyboard",
            "quantity": 1,
            "price": {"amount": 89.99, "currency": "USD"},
            "sku": "KB-MECH-001",
            "hs_code": "8471.60.80",
            "origin_country": "CN"
          }
        ]
      }
    ]
  }
}
EOF
```

**注意：** 替换 `account_id` 为实际的账户 ID

#### 执行命令（wait=10 参数）

```bash
curl -X POST "http://localhost:8080/api/v1/orders?wait=10" \
  -H "Content-Type: application/json" \
  -d @/tmp/test_order_smart_wait.json | jq .
```

**参数说明：**

- `wait=10`：API 会 hold 请求最多 10 秒，等待诊断完成
- 如果 10 秒内完成，返回 200 + 完整诊断结果
- 如果超时，返回 3001 + order_id

#### 预期响应（200 OK + 完整诊断结果）

```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "id": "2a5fe3fc-d5c3-46b0-91fd-a7960c5091a2",
    "account_id": 6265158001000,
    "merchant_order_no": "TEST-ORDER-SMART-WAIT-001",
    "status": "DIAGNOSED",
    "diagnosis": {
      "items": [
        {
          "type": "shipping",
          "status": "SUCCESS",
          "data_json": {
            "rates": [
              {
                "carrier": "USPS",
                "service": "Priority Mail",
                "cost": 15.99,
                "currency": "USD",
                "delivery_days": 3
              },
              {
                "carrier": "UPS",
                "service": "Ground",
                "cost": 18.50,
                "currency": "USD",
                "delivery_days": 5
              }
            ],
            "recommended": {
              "carrier": "USPS",
              "service": "Priority Mail",
              "reason": "Lowest cost with acceptable delivery time"
            }
          }
        },
        {
          "type": "anomaly",
          "status": "SUCCESS",
          "data_json": {
            "checks": [
              {
                "type": "address_validation",
                "status": "PASS",
                "message": "Address is valid"
              },
              {
                "type": "weight_dimension_check",
                "status": "PASS",
                "message": "Weight and dimensions are reasonable"
              },
              {
                "type": "customs_check",
                "status": "PASS",
                "message": "HS code is valid for declared items"
              }
            ],
            "risk_level": "LOW",
            "anomalies_found": []
          }
        }
      ]
    },
    "created_at": "2025-01-09T10:05:00+08:00",
    "updated_at": "2025-01-09T10:05:03+08:00"
  }
}
```

#### 诊断结果结构说明

**diagnosis.items[] 数组：**

每个诊断项包含：
- `type`：诊断类型（`shipping` 或 `anomaly`）
- `status`：诊断状态（`SUCCESS`, `FAILED`, `SKIPPED`）
- `data_json`：具体诊断数据（结构因 type 而异）

**shipping 诊断数据：**
```json
{
  "rates": [...]           // 费率列表
  "recommended": { ... }   // 推荐方案
}
```

**anomaly 诊断数据：**
```json
{
  "checks": [...]          // 检查项列表
  "risk_level": "LOW",     // 风险等级: LOW/MEDIUM/HIGH
  "anomalies_found": []    // 发现的异常
}
```

#### 完整链路验证说明（10 个步骤）

当收到 200 响应时，说明以下链路全部成功：

```
✓ 步骤 1：[dpmain] 接收订单创建请求
✓ 步骤 2：[dpmain] 保存订单到 MySQL
✓ 步骤 3：[dpmain] 推送诊断任务到 Lmstfy 队列 (order_diagnose)
✓ 步骤 4：[dpsync] 从队列消费诊断任务
✓ 步骤 5：[dpsync] 执行诊断逻辑 (shipping 费率计算)
✓ 步骤 6：[dpsync] 执行诊断逻辑 (anomaly 异常检测)
✓ 步骤 7：[dpsync] 推送回调到 callback 队列 (order_diagnose_callback)
✓ 步骤 8：[dpmain] Callback Consumer 消费回调
✓ 步骤 9：[dpmain] 更新订单状态和诊断结果到 MySQL
✓ 步骤 10：[dpmain] 通过 Redis Pub/Sub 通知等待的 API 请求
✓ 步骤 11：[dpmain] API 返回完整诊断结果给客户端
```

#### 同时观察两个服务的日志输出

**终端 1 - dpmain 日志：**
```
[INFO] Received order request: merchant_order_no=TEST-ORDER-SMART-WAIT-001
[INFO] Order created: id=2a5fe3fc-d5c3-46b0-91fd-a7960c5091a2
[INFO] Published diagnosis job to Lmstfy: job_id=01KDCBQQWTXBJ2ZBYP9W000000
[INFO] Waiting for diagnosis result... (max 10s)
[INFO] Received callback message: order_id=2a5fe3fc-d5c3-46b0-91fd-a7960c5091a2
[INFO] Order updated with diagnosis result
[INFO] Published result to Redis channel: diagnosis:result:2a5fe3fc-d5c3-46b0-91fd-a7960c5091a2
[INFO] Smart Wait: Received result after 2.8s
[INFO] Response sent: 200 OK
```

**终端 2 - dpsync 日志：**
```
[INFO] Consumed message: job_id=01KDCBQQWTXBJ2ZBYP9W000000
[INFO] Processing diagnosis for order: 2a5fe3fc-d5c3-46b0-91fd-a7960c5091a2
[INFO] Executing shipping calculator...
[INFO] Calculated 2 shipping rates
[INFO] Executing anomaly checker...
[INFO] Completed 3 anomaly checks, risk_level=LOW
[INFO] Diagnosis completed successfully
[INFO] Published callback to Lmstfy: queue=order_diagnose_callback
[INFO] ACK message: job_id=01KDCBQQWTXBJ2ZBYP9W000000
```

**保存订单 ID 到环境变量：**
```bash
export ORDER_ID="2a5fe3fc-d5c3-46b0-91fd-a7960c5091a2"  # 替换为实际返回的 ID
```

---

### 4.5 测试 5：查询订单详情

#### 执行命令

```bash
curl http://localhost:8080/api/v1/orders/$ORDER_ID | jq .
```

#### 预期响应（包含诊断结果）

```json
{
  "code": 200,
  "message": "Success",
  "data": {
    "id": "2a5fe3fc-d5c3-46b0-91fd-a7960c5091a2",
    "account_id": 6265158001000,
    "merchant_order_no": "TEST-ORDER-SMART-WAIT-001",
    "status": "DIAGNOSED",
    "shipment": {
      "ship_from": { ... },
      "ship_to": { ... },
      "parcels": [ ... ]
    },
    "diagnosis": {
      "items": [
        {
          "type": "shipping",
          "status": "SUCCESS",
          "data_json": { ... }
        },
        {
          "type": "anomaly",
          "status": "SUCCESS",
          "data_json": { ... }
        }
      ]
    },
    "created_at": "2025-01-09T10:05:00+08:00",
    "updated_at": "2025-01-09T10:05:03+08:00"
  }
}
```

---

### 4.6 测试 6：验证诊断结果结构

使用 jq 提取和验证诊断结果：

#### 检查诊断项目数量

```bash
curl -s http://localhost:8080/api/v1/orders/$ORDER_ID | \
  jq '.data.diagnosis.items | length'
```

**预期输出：**
```
2
```

#### 检查诊断项目类型

```bash
curl -s http://localhost:8080/api/v1/orders/$ORDER_ID | \
  jq '.data.diagnosis.items[].type'
```

**预期输出：**
```
"shipping"
"anomaly"
```

#### 检查 shipping 费率数量

```bash
curl -s http://localhost:8080/api/v1/orders/$ORDER_ID | \
  jq '.data.diagnosis.items[] | select(.type=="shipping") | .data_json.rates | length'
```

**预期输出：**
```
2
```

#### 检查推荐承运商

```bash
curl -s http://localhost:8080/api/v1/orders/$ORDER_ID | \
  jq '.data.diagnosis.items[] | select(.type=="shipping") | .data_json.recommended.carrier'
```

**预期输出：**
```
"USPS"
```

#### 检查异常风险等级

```bash
curl -s http://localhost:8080/api/v1/orders/$ORDER_ID | \
  jq '.data.diagnosis.items[] | select(.type=="anomaly") | .data_json.risk_level'
```

**预期输出：**
```
"LOW"
```

---

## 5. 查看和分析日志

### 5.1 实时监控日志

#### 查看 dpmain 日志（终端 4）

```bash
# 如果 dpmain 在前台运行，直接查看终端 1
# 如果在后台运行，使用：
tail -f /tmp/dpmain-apiserver.log
```

#### 查看 dpsync 日志（终端 5）

```bash
# 如果 dpsync 在前台运行，直接查看终端 2
# 如果在后台运行，使用：
tail -f /tmp/dpsync-worker.log
```

### 5.2 关键日志模式

#### 成功的诊断流程日志

**dpmain 关键日志：**
```
[INFO] Published diagnosis job to Lmstfy: job_id=...
[INFO] Waiting for diagnosis result...
[INFO] Received callback message: order_id=...
[INFO] Smart Wait: Received result after X.Xs
```

**dpsync 关键日志：**
```
[INFO] Consumed message: job_id=...
[INFO] Processing diagnosis for order: ...
[INFO] Diagnosis completed successfully
[INFO] Published callback to Lmstfy
[INFO] ACK message: job_id=...
```

#### 消息队列消费日志

```
[DEBUG] Polling Lmstfy: queue=order_diagnose, timeout=5s
[INFO] Consumed message: job_id=..., data=...
[INFO] ACK message: job_id=...
```

#### Redis Pub/Sub 日志

```
[DEBUG] Publishing to Redis channel: diagnosis:result:...
[DEBUG] Subscribed to Redis channel: diagnosis:result:...
[DEBUG] Received message from Redis channel
```

---

## 6. 清理测试环境

### 停止应用服务

在启动服务的终端按 `Ctrl+C` 停止：

**终端 1 - dpmain：**
```bash
# Ctrl+C
^C
[INFO] Received interrupt signal
[INFO] Gracefully shutting down...
[INFO] HTTP server stopped
[INFO] Callback consumer stopped
```

**终端 2 - dpsync：**
```bash
# Ctrl+C
^C
[INFO] Received interrupt signal
[INFO] Worker stopped gracefully
```

### 清理测试数据（可选）

```bash
# 删除测试订单
docker exec -i oip_mysql mysql -uroot -ppassword oip \
  -e "DELETE FROM orders WHERE merchant_order_no LIKE 'TEST-ORDER-%';"

# 删除测试账户
docker exec -i oip_mysql mysql -uroot -ppassword oip \
  -e "DELETE FROM accounts WHERE email LIKE '%@example.com';"
```

### 保留或清理 Docker 容器

**保留容器（推荐）：**
```bash
# 容器继续运行，方便下次测试
docker-compose ps
```

**停止容器：**
```bash
docker-compose stop
```

**完全清理（删除数据）：**
```bash
docker-compose down -v
```

---

# 第二部分：自动化 E2E 测试

## 7. 使用 verify.sh 一键测试

### 7.1 脚本说明

`verify.sh` 脚本提供一键自动化测试，包含：
- ✅ 模块构建验证
- ✅ 自动启动服务（dpmain + dpsync）
- ✅ 健康检查
- ✅ E2E 测试执行（账户 + 订单 + 诊断）
- ✅ 自动清理（停止服务）

### 7.2 执行命令

```bash
cd /Users/cooperswang/Documents/wsqyouth/oip/backend
./scripts/verify.sh
```

### 7.3 执行流程

```
1. 检查 Go 版本
2. 检查当前目录
3. 清理模块缓存（已禁用）
4. 验证 common 模块
5. 验证并构建 dpmain 模块
6. 验证并构建 dpsync 模块
7. 验证 Go Workspace
8. 启动测试服务
   ├─ 检查端口占用
   ├─ 检查 Docker 服务
   ├─ 启动 dpmain
   ├─ 启动 dpsync
   └─ 健康检查
9. E2E 测试
   ├─ 步骤 1/4: 创建测试账户
   ├─ 步骤 2/4: 验证账户查询
   ├─ 步骤 3/4: 创建订单并等待诊断（Smart Wait 15s）
   └─ 步骤 4/4: 验证订单查询
```

**预期执行时间：** 约 1-2 分钟

### 7.4 成功输出示例

```bash
=========================================
  OIP Backend - 验证构建脚本
=========================================

1. 检查 Go 版本...
go version go1.21.5 darwin/arm64

...

8. 启动测试服务...
   ✓ 端口 8080 可用
   ✓ Docker 服务运行正常 (MySQL, Redis, Lmstfy)
   ✓ dpmain 已启动 (PID: 12345)
   ✓ dpmain 服务就绪 (耗时: 3s)
   ✓ dpsync 已启动 (PID: 12346)
   ✓ dpsync 服务就绪

   ✓✓✓ 所有服务启动成功！
       - dpmain:  http://localhost:8080 (PID: 12345)
       - dpsync:  Worker 运行中 (PID: 12346)

   日志位置:
       - dpmain: /tmp/oip-verify-logs/dpmain.log
       - dpsync: /tmp/oip-verify-logs/dpsync.log

9. E2E 测试：订单创建与诊断完整链路...

   -> 步骤 1/4: 创建测试账户...
   ✓ 账户创建成功 (ID: 6265158001000)
   -> 步骤 2/4: 验证账户查询...
   ✓ 账户查询成功
   -> 步骤 3/4: 创建订单并等待诊断 (Smart Wait 15秒)...
   ✓ 订单创建成功 (ID: 2a5fe3fc-d5c3-46b0-91fd-a7960c5091a2)
   ✓ 诊断已完成 (状态: DIAGNOSED)
   ✓ 诊断结果包含 2 个项目
   ✓ 诊断类型: [shipping, anomaly]
   -> 步骤 4/4: 验证订单查询...
   ✓ 订单查询成功

   ✓✓✓ E2E 测试完整链路验证成功！

   验证的完整链路：
       1. [dpmain] 接收订单创建请求 ✓
       2. [dpmain] 保存订单到 MySQL ✓
       3. [dpmain] 推送诊断任务到 Lmstfy 队列 (order_diagnose) ✓
       4. [dpsync] 从队列消费诊断任务 ✓
       5. [dpsync] 执行诊断逻辑 (shipping, anomaly) ✓
       6. [dpsync] 推送回调到 callback 队列 (order_diagnose_callback) ✓
       7. [dpmain] Callback Consumer 消费回调 ✓
       8. [dpmain] 更新订单状态和诊断结果到 MySQL ✓
       9. [dpmain] 通过 Redis Pub/Sub 通知等待的 API 请求 ✓
      10. [dpmain] API 返回完整诊断结果给客户端 ✓

   测试数据：
       - 账户ID: 6265158001000
       - 订单ID: 2a5fe3fc-d5c3-46b0-91fd-a7960c5091a2
       - 订单号: E2E-TEST-1736395200
       - 诊断项: 2 个 (shipping, anomaly)

=========================================
  ✓✓✓ 所有测试通过！
=========================================

日志文件位置：
  - dpmain: /tmp/oip-verify-logs/dpmain.log
  - dpsync: /tmp/oip-verify-logs/dpsync.log

提示: 服务将在脚本退出时自动停止

=========================================
  清理测试环境...
=========================================
   -> 停止 dpmain 服务 (PID: 12345)...
   ✓ dpmain 服务已停止
   -> 停止 dpsync 服务 (PID: 12346)...
   ✓ dpsync 服务已停止
   ✓ 清理完成

日志文件位置：
   - dpmain: /tmp/oip-verify-logs/dpmain.log
   - dpsync: /tmp/oip-verify-logs/dpsync.log
```

### 7.5 诊断超时场景（3001）

如果诊断在 15 秒内未完成，脚本会输出：

```
   ⚠️  订单创建成功但诊断超时 (ID: xxx, Code: 3001)

   这不是预期结果。检查项：
       - dpsync 服务是否正常运行？
       - callback consumer 是否正常消费？
       - Lmstfy 队列是否正常？

   检查 dpsync 日志:
   [最近30行日志...]

   检查 dpmain 日志:
   [最近30行日志...]
```

### 7.6 查看日志文件

所有日志保存在 `/tmp/oip-verify-logs/`：

```bash
# 查看 dpmain 完整日志
cat /tmp/oip-verify-logs/dpmain.log

# 查看 dpsync 完整日志
cat /tmp/oip-verify-logs/dpsync.log

# 实时监控（脚本运行时）
tail -f /tmp/oip-verify-logs/dpmain.log
tail -f /tmp/oip-verify-logs/dpsync.log
```

### 7.7 修改测试数据

脚本中的订单测试数据定义在 `ORDER_PAYLOAD` 变量中，可以根据需要修改：

**编辑脚本：**
```bash
vi /Users/cooperswang/Documents/wsqyouth/oip/backend/scripts/verify.sh
```

**找到并修改 ORDER_PAYLOAD 部分：**
```bash
# 定义订单测试数据（可根据需要修改）
ORDER_PAYLOAD=$(cat <<EOF
{
    "account_id": $ACCOUNT_ID,
    "merchant_order_no": "$ORDER_NO",
    "shipment": {
        "ship_from": {
            "contact_name": "E2E Test Store",
            "street1": "230 W 200 S",
            "city": "Salt Lake City",
            "state": "UT",
            "postal_code": "84101",
            "country": "US",
            "phone": "+1-801-555-0100",
            "email": "store@e2etest.com"
        },
        # ... 修改这里的测试数据 ...
    }
}
EOF
)
```

---

## 8. 故障排查

### 8.1 环境问题

#### 问题：Docker 未启动

**错误信息：**
```
✗ Docker 未运行，请先启动 Docker
```

**解决方法：**
```bash
# macOS
open -a Docker

# Linux
sudo systemctl start docker
```

#### 问题：端口被占用（8080）

**错误信息：**
```
✗ 端口 8080 已被占用，请先停止占用该端口的进程
```

**解决方法：**
```bash
# 查找占用进程
lsof -i :8080

# 杀掉进程
kill -9 <PID>
```

#### 问题：依赖服务未启动

**错误信息：**
```
⚠️  警告: Docker 服务未完全启动 (当前: 1/3)
```

**解决方法：**
```bash
# 启动所有依赖服务
docker-compose up -d

# 等待启动完成
sleep 5

# 验证
docker-compose ps
```

### 8.2 服务启动失败

#### 问题：dpmain 启动超时

**错误信息：**
```
✗ dpmain 服务启动超时
```

**排查步骤：**

1. 查看日志：
```bash
tail -50 /tmp/oip-verify-logs/dpmain.log
```

2. 检查配置文件：
```bash
cat /Users/cooperswang/Documents/wsqyouth/oip/backend/dpmain/config/config.yaml
```

3. 检查 MySQL 连接：
```bash
docker exec -i oip_mysql mysql -uroot -ppassword -e "SELECT 1;"
```

4. 检查 Redis 连接：
```bash
docker exec -i oip_redis redis-cli ping
```

#### 问题：dpsync 启动失败

**错误信息：**
```
✗ dpsync 服务启动失败
```

**排查步骤：**

1. 查看日志：
```bash
tail -50 /tmp/oip-verify-logs/dpsync.log
```

2. 检查 Lmstfy 连接：
```bash
curl http://localhost:7777/ping
```

3. 验证配置：
```bash
cat /Users/cooperswang/Documents/wsqyouth/oip/backend/dpsync/config/worker.yaml
```

#### 问题：dpsync Lmstfy 认证失败（401 invalid token）

**错误信息：**
```
{"level":"warn","msg":"[Subscriber-0] Consume error: lmstfy consume failed: t:resp; m:[401]invalid token; j:; r:..."}
```

**问题根因：**

dpsync 和 dpmain 使用了不同的 Lmstfy token：

- ❌ **dpsync（错误）**: `01KDCBF5BG0THBC24F1V53XPR1`
- ✅ **dpmain（正确）**: `01KEED5FWJB9GT21S0GVQ4SXXW`

**验证方法：**

```bash
# 测试 dpmain 的 token - 有效 ✓
curl -s -H "X-Token: 01KEED5FWJB9GT21S0GVQ4SXXW" \
  "http://localhost:7777/api/oip/order_diagnose?timeout=1&ttr=30"
# 返回: {"job_id":"...","msg":"new job",...}

# 测试 dpsync 的旧 token - 无效 ✗
curl -s -H "X-Token: 01KDCBF5BG0THBC24F1V53XPR1" \
  "http://localhost:7777/api/oip/order_diagnose?timeout=1&ttr=30"
# 返回: {"error":"invalid token"}
```

**解决方法：**

1. 编辑 dpsync 配置文件：
```bash
vi /Users/cooperswang/Documents/wsqyouth/oip/backend/dpsync/config/worker.yaml
```

2. 修改 token 为正确值：
```yaml
lmstfy:
  host: "localhost"
  port: 7777
  namespace: "oip"
  token: "01KEED5FWJB9GT21S0GVQ4SXXW"  # 改为与 dpmain 一致
```

3. 重启 dpsync 服务：
```bash
cd /Users/cooperswang/Documents/wsqyouth/oip/backend/dpsync
# 停止旧进程
pkill -f dpsync-worker

# 重新启动
./bin/dpsync-worker
```

**验证修复：**

正常启动日志应显示：
```
{"level":"info","msg":"[Manager] Starting..."}
{"level":"info","msg":"[Worker] order-diagnose-worker started"}
{"level":"info","msg":"[Subscriber] Started with 3 threads"}
Worker started. Press Ctrl+C to shutdown.
```

#### 问题：配置文件错误

**常见错误：**
- 数据库连接字符串错误
- Redis 地址错误
- Lmstfy 地址或 token 错误

**验证配置：**
```bash
# 检查 dpmain 配置
cat dpmain/config/config.yaml

# 检查 dpsync 配置
cat dpsync/config/config.yaml
```

### 8.3 测试失败场景

#### 问题：账户创建失败

**错误信息：**
```
✗ 创建账户失败
```

**可能原因：**
1. 数据库未初始化（accounts 表不存在）
2. 数据库连接失败
3. 参数验证失败

**解决方法：**
```bash
# 检查表是否存在
docker exec -i oip_mysql mysql -uroot -ppassword oip \
  -e "SHOW TABLES;" 2>&1 | grep accounts

# 如果不存在，运行迁移脚本
cd dpmain
# 运行数据库迁移...
```

#### 问题：订单创建失败

**错误信息：**
```
✗ 订单创建失败 (Code: 400)
```

**可能原因：**
1. account_id 无效
2. 订单数据格式错误
3. 必填字段缺失

**调试方法：**
```bash
# 查看完整响应
curl -X POST "http://localhost:8080/api/v1/orders?wait=10" \
  -H "Content-Type: application/json" \
  -d @/tmp/test_order_smart_wait.json \
  -v

# 检查 dpmain 日志
tail -30 /tmp/oip-verify-logs/dpmain.log
```

#### 问题：诊断超时（3001）

**现象：**
- 订单创建成功
- 但在 wait 时间内未收到诊断结果
- 返回 Code 3001

**可能原因：**
1. dpsync 服务未运行
2. dpsync 消费消息失败
3. callback consumer 未正常消费
4. Redis Pub/Sub 连接问题
5. 诊断逻辑执行时间过长

**排查步骤：**

1. 检查 dpsync 是否运行：
```bash
ps aux | grep dpsync-worker
```

2. 查看 dpsync 日志：
```bash
tail -50 /tmp/oip-verify-logs/dpsync.log | grep -i error
```

3. 检查消息队列：
```bash
curl -s -H "X-Token: 01KEED5FWJB9GT21S0GVQ4SXXW" \
  "http://localhost:7777/api/oip/order_diagnose?timeout=1&ttr=30"
```

4. 检查 Redis Pub/Sub：
```bash
# 终端 1
docker exec -i oip_redis redis-cli SUBSCRIBE "diagnosis:result:*"

# 终端 2（创建订单）
curl -X POST "http://localhost:8080/api/v1/orders?wait=10" \
  -H "Content-Type: application/json" \
  -d @/tmp/test_order_smart_wait.json
```

5. 查看 dpmain callback consumer 日志：
```bash
grep "Received callback" /tmp/oip-verify-logs/dpmain.log
```

#### 问题：诊断结果为空

**错误信息：**
```
✗ 诊断结果为空
```

**可能原因：**
1. dpsync 诊断逻辑错误
2. 回调数据序列化失败
3. 数据库更新失败

**排查步骤：**
```bash
# 查看 dpsync 日志中的诊断过程
grep "diagnosis" /tmp/oip-verify-logs/dpsync.log -i

# 直接查询数据库
docker exec -i oip_mysql mysql -uroot -ppassword oip \
  -e "SELECT id, status, diagnose_result FROM orders WHERE id='$ORDER_ID';" \
  2>&1 | grep -v Warning
```

### 8.4 日志分析指南

#### dpmain 关键日志点

**正常流程：**
```
[INFO] Received order request             # 收到请求
[INFO] Order created: id=...              # 订单创建
[INFO] Published diagnosis job            # 推送任务
[INFO] Waiting for diagnosis result       # 等待结果
[INFO] Received callback message          # 收到回调
[INFO] Order updated with diagnosis       # 更新订单
[INFO] Published result to Redis          # 发布通知
[INFO] Smart Wait: Received result        # 返回结果
```

**错误日志：**
```
[ERROR] Failed to create order            # 订单创建失败
[ERROR] Failed to publish to Lmstfy       # 队列推送失败
[ERROR] Failed to update order            # 更新失败
[WARN]  Smart Wait timeout                # 等待超时
```

#### dpsync 关键日志点

**正常流程：**
```
[INFO] Consumed message: job_id=...      # 消费消息
[INFO] Processing diagnosis for order    # 开始诊断
[INFO] Executing shipping calculator     # 费率计算
[INFO] Executing anomaly checker         # 异常检测
[INFO] Diagnosis completed successfully  # 诊断完成
[INFO] Published callback to Lmstfy      # 推送回调
[INFO] ACK message                        # 确认消息
```

**错误日志：**
```
[ERROR] Failed to consume message         # 消费失败
[ERROR] Diagnosis failed                  # 诊断失败
[ERROR] Failed to publish callback        # 回调推送失败
```

#### 消息队列问题排查

**检查队列状态：**
```bash
# 查看 order_diagnose 队列
curl -s -H "X-Token: 01KEED5FWJB9GT21S0GVQ4SXXW" \
  "http://localhost:7777/api/oip/order_diagnose?timeout=1&ttr=30"

# 查看 order_diagnose_callback 队列
curl -s -H "X-Token: 01KEED5FWJB9GT21S0GVQ4SXXW" \
  "http://localhost:7777/api/oip/order_diagnose_callback?timeout=1&ttr=30"
```

**查看 Lmstfy 日志：**
```bash
docker logs oip_lmstfy --tail 50
```

#### Redis 连接问题

**测试 Redis 连接：**
```bash
docker exec -i oip_redis redis-cli ping
```

**测试 Pub/Sub：**
```bash
# 终端 1：订阅测试
docker exec -i oip_redis redis-cli SUBSCRIBE test_channel

# 终端 2：发布测试
docker exec -i oip_redis redis-cli PUBLISH test_channel "hello"
```

**查看 Redis 日志：**
```bash
docker logs oip_redis --tail 50
```

---

## 9. 验证清单

### 9.1 手动测试清单

- [ ] **环境准备**
  - [ ] Docker 服务运行正常
  - [ ] MySQL 容器运行（端口 3306）
  - [ ] Redis 容器运行（端口 6379）
  - [ ] Lmstfy 容器运行（端口 7777）
  - [ ] 依赖服务连通性测试通过

- [ ] **服务启动**
  - [ ] dpmain 构建成功
  - [ ] dpsync 构建成功
  - [ ] dpmain 服务启动成功（端口 8080）
  - [ ] dpsync 服务启动成功
  - [ ] 健康检查接口返回 200

- [ ] **功能测试**
  - [ ] 账户创建成功（Code 200）
  - [ ] 账户查询成功
  - [ ] 订单创建成功（不等待，Code 3001）
  - [ ] 订单创建成功（Smart Wait，Code 200）
  - [ ] 诊断结果包含 2 个项目（shipping, anomaly）
  - [ ] 订单查询返回完整数据

- [ ] **链路验证**
  - [ ] dpmain 日志显示推送任务
  - [ ] dpsync 日志显示消费消息
  - [ ] dpsync 日志显示诊断完成
  - [ ] dpmain 日志显示收到回调
  - [ ] dpmain 日志显示 Smart Wait 返回结果

- [ ] **数据验证**
  - [ ] 数据库中订单状态为 DIAGNOSED
  - [ ] 诊断结果 JSON 结构正确
  - [ ] shipping 包含费率列表
  - [ ] anomaly 包含检查结果

### 9.2 自动化测试清单

- [ ] **脚本执行**
  - [ ] verify.sh 执行无错误
  - [ ] 所有 9 个步骤通过
  - [ ] 服务自动启动成功
  - [ ] E2E 测试 4 个子步骤全部通过

- [ ] **链路验证**
  - [ ] 10 个链路步骤验证通过
  - [ ] 诊断结果包含预期数据
  - [ ] 日志文件生成正常

- [ ] **清理验证**
  - [ ] 服务自动停止
  - [ ] 日志文件可访问
  - [ ] 无僵尸进程遗留

---

## 10. 附录

### 10.1 常用调试命令

#### Docker 相关

```bash
# 查看所有容器
docker ps -a

# 查看容器日志
docker logs oip_mysql --tail 50
docker logs oip_redis --tail 50
docker logs oip_lmstfy --tail 50

# 重启容器
docker restart oip_mysql
docker restart oip_redis
docker restart oip_lmstfy

# 进入容器
docker exec -it oip_mysql bash
docker exec -it oip_redis sh
```

#### 数据库查询

```bash
# 查看所有账户
docker exec -i oip_mysql mysql -uroot -ppassword oip \
  -e "SELECT * FROM accounts ORDER BY id DESC LIMIT 10;" \
  2>&1 | grep -v Warning

# 查看所有订单
docker exec -i oip_mysql mysql -uroot -ppassword oip \
  -e "SELECT id, merchant_order_no, status, created_at FROM orders ORDER BY created_at DESC LIMIT 10;" \
  2>&1 | grep -v Warning

# 查看订单详情
docker exec -i oip_mysql mysql -uroot -ppassword oip \
  -e "SELECT * FROM orders WHERE id='$ORDER_ID'\G" \
  2>&1 | grep -v Warning

# 查看诊断结果 JSON
docker exec -i oip_mysql mysql -uroot -ppassword oip \
  -e "SELECT diagnose_result FROM orders WHERE id='$ORDER_ID'\G" \
  2>&1 | grep -v Warning
```

#### Redis 命令

```bash
# 连接 Redis
docker exec -it oip_redis redis-cli

# 查看所有 keys
docker exec -i oip_redis redis-cli KEYS "*"

# 订阅 channel（用于调试）
docker exec -i oip_redis redis-cli SUBSCRIBE "diagnosis:result:*"

# 查看 Redis 信息
docker exec -i oip_redis redis-cli INFO
```

#### Lmstfy 队列查询

```bash
# 查看队列消息（需要 token）
TOKEN="01KEED5FWJB9GT21S0GVQ4SXXW"

# 查看 order_diagnose 队列
curl -s -H "X-Token: $TOKEN" \
  "http://localhost:7777/api/oip/order_diagnose?timeout=1&ttr=30" | jq .

# 查看 order_diagnose_callback 队列
curl -s -H "X-Token: $TOKEN" \
  "http://localhost:7777/api/oip/order_diagnose_callback?timeout=1&ttr=30" | jq .

# 查看队列统计
curl -s "http://localhost:7777/api/oip/order_diagnose/stats" | jq .
```

### 10.2 性能基准

#### 预期响应时间

| 操作 | 预期时间 |
|------|----------|
| 创建账户 | < 100ms |
| 查询账户 | < 50ms |
| 创建订单（不等待） | < 200ms |
| 创建订单（Smart Wait） | 2-5s |
| 查询订单 | < 100ms |

#### 诊断完成时间范围

| 诊断类型 | 预期时间 |
|----------|----------|
| Shipping 费率计算 | 500ms - 2s |
| Anomaly 异常检测 | 300ms - 1s |
| 完整诊断流程 | 2s - 5s |

**注意：** 实际时间受以下因素影响：
- 网络延迟
- 队列消费延迟
- 第三方 API 调用（如有）
- 系统负载

### 10.3 参考文档链接

- **项目文档**
  - [PRD.md](../PRD.md) - 产品需求文档
  - [ARCHITECTURE.md](../ARCHITECTURE.md) - 架构设计
  - [dpmain/README.md](../dpmain/README.md) - dpmain 服务说明
  - [dpsync/README.md](../dpsync/README.md) - dpsync 服务说明

- **测试相关**
  - [TESTING_GUIDE.md](../dpmain/TESTING_GUIDE.md) - 测试指南
  - [scripts/verify.sh](../scripts/verify.sh) - 自动化验证脚本

- **开发相关**
  - [.claude.md](../.claude.md) - Claude 上下文
  - [stories/](../stories/) - 开发任务列表

---

## 总结

本文档提供了 OIP Backend 完整的 E2E 测试流程，包括：

1. **手动测试**：详细的步骤说明，适合学习和调试
2. **自动化测试**：一键验证脚本，适合持续集成
3. **故障排查**：常见问题和解决方案
4. **调试工具**：实用的命令和技巧

**推荐测试流程：**
1. 首次测试：按照手动测试步骤逐步执行，理解完整链路
2. 日常开发：使用 `verify.sh` 快速验证
3. 遇到问题：参考故障排查章节，查看日志分析

**联系方式：**
如有问题，请查看日志文件或联系开发团队。

---

**文档版本：** v1.0
**最后更新：** 2025-01-09
**维护者：** OIP Backend Team
