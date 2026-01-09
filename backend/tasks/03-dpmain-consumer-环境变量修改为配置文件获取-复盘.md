# 03-dpmain consumer 环境变量修改为配置文件获取 - 复盘与 Lmstfy 深度学习

## 一、问题与解决方案

### 问题现象
```
[DPMAIN] 2026/01/08 16:52:32 [ERROR] Failed to consume message
%!(EXTRA string=error, *fmt.wrapError=consume message failed: lmstfy consume failed: status=401)
```
dpmain 启动后持续报 401 认证失败，无法消费回调队列。

### 解决方案
**根本原因**：代码中硬编码的 lmstfy token `01KDCBF5BG0THBC24F1V53XPR1` 未在 lmstfy 服务器注册。

**修复步骤**：
1. 通过 lmstfy 管理 API 创建有效 token
   ```bash
   curl -XPOST -d "description=OIP Backend Service" "http://localhost:7778/token/oip"
   # 响应: {"token": "01KEED5FWJB9GT21S0GVQ4SXXW"}
   ```

2. 重构配置管理：环境变量 → 配置文件
   - 创建 `config/config.yaml`
   - 使用 viper 读取配置
   - 修复两个入口：`apiserver` 和 `callback_consumer`

**建议**：后续将敏感配置迁移到配置中心（K8s ConfigMap/Secret）或 Vault。

---

## 二、Lmstfy 深度学习

### 2.1 Token 管理

#### 原型定义
```
POST /token/:namespace          # 创建 token
GET  /token/:namespace          # 查询 token 列表
DELETE /token/:namespace/:token # 删除 token
```

#### 实战示例

**1. 创建 namespace 和 token**
```bash
curl -XPOST -d "description=OIP Backend Service" \
  "http://127.0.0.1:7778/token/oip"
```

**响应：**
```json
{
  "token": "01KEED5FWJB9GT21S0GVQ4SXXW"
}
```

**2. 查询 namespace 下所有 token**
```bash
curl "http://127.0.0.1:7778/token/oip"
```

**响应：**
```json
{
  "tokens": {
    "01KEED5FWJB9GT21S0GVQ4SXXW": "OIP Backend Service"
  }
}
```

**3. 删除 token**
```bash
curl -XDELETE "http://127.0.0.1:7778/token/oip/01KEED5FWJB9GT21S0GVQ4SXXW"
```

**响应：**
```
204 No Content
```

#### Redis 数据检查

Token 存储在 Redis 中，可以直接查看：

```bash
# 1. 进入 Redis 容器
docker exec -it oip_redis redis-cli

# 2. 查看所有 token 相关键
KEYS *token*

# 3. 查看 namespace 的 token 数据
HGETALL "pool:default/namespace:oip/tokens"

# 示例输出：
# 1) "01KEED5FWJB9GT21S0GVQ4SXXW"
# 2) "OIP Backend Service"
```

**Token 特性：**
- ✅ 永久有效（无过期时间）
- ✅ 持久化存储（存储在 Redis，重启不丢失）
- ✅ 手动管理（需主动删除才失效）

---

### 2.2 Job 生命周期管理

#### 原型定义
```
PUT    /api/:namespace/:queue?delay=&ttl=&tries=  # 发布 job
GET    /api/:namespace/:queue?timeout=&ttr=       # 消费 job
DELETE /api/:namespace/:queue/job/:job_id         # ACK job
GET    /api/:namespace/:queue/peek                # Peek job（不消费）
```

#### 参数说明

**发布参数：**
- `delay`: 延迟执行时间（秒），0 表示立即可消费
- `ttl` (Time-To-Live): Job 存活时间（秒），0 表示永不过期
- `tries`: 最大重试次数（默认 1）

**消费参数：**
- `timeout`: 等待超时（秒），队列为空时最多等待多久（long polling）
- `ttr` (Time-To-Run): 处理超时（秒），超时未 ACK 则重新入队

#### 实战示例

**1. 发布 Job**
```bash
# 创建一个 job: delay=1s, ttl=3600s, tries=3
curl -XPUT \
  -H "X-Token:01KEED5FWJB9GT21S0GVQ4SXXW" \
  -d '{"order_id": "TEST001", "action": "diagnose"}' \
  "http://127.0.0.1:7777/api/oip/test_queue?delay=1&ttl=3600&tries=3"
```

**响应：**
```json
{
  "job_id": "01KEEH8F7Z9XXXXXXXXXXXXXXX"
}
```

**2. 消费 Job**
```bash
curl -H "X-Token:01KEED5FWJB9GT21S0GVQ4SXXW" \
  "http://127.0.0.1:7777/api/oip/test_queue?ttr=30&timeout=2"
```

**响应：**
```json
{
  "msg": "new job",
  "namespace": "oip",
  "queue": "test_queue",
  "job_id": "01KEEH8F7Z9XXXXXXXXXXXXXXX",
  "data": "eyJvcmRlcl9pZCI6ICJURVNUMDA...",  // base64 编码
  "ttl": 3599,
  "elapsed_ms": 1234
}
```

**3. ACK Job（确认消费）**
```bash
curl -i -XDELETE \
  -H "X-Token:01KEED5FWJB9GT21S0GVQ4SXXW" \
  "http://127.0.0.1:7777/api/oip/test_queue/job/01KEEH8F7Z9XXXXXXXXXXXXXXX"
```

**响应：**
```
HTTP/1.1 204 No Content
```

**4. Peek Job（查看但不消费）**
```bash
curl -H "X-Token:01KEED5FWJB9GT21S0GVQ4SXXW" \
  "http://127.0.0.1:7777/api/oip/test_queue/peek"
```

**响应：**
```json
{
  "namespace": "oip",
  "queue": "test_queue",
  "job_id": "01KEEH8F7Z9XXXXXXXXXXXXXXX",
  "data": "eyJvcmRlcl9pZCI6ICJURVNUMDA...",
  "ttl": 3599
}
```

**注意**：Peek 无法返回 TTL 已过期的 job 数据。

#### Redis 数据检查

**Job 在不同阶段的 Redis 键：**

```bash
# 1. 查看 Ready 队列（等待消费）
redis-cli LLEN "pool:default/namespace:oip/q:test_queue"
redis-cli LRANGE "pool:default/namespace:oip/q:test_queue" 0 -1

# 2. 查看 Job 数据
redis-cli GET "pool:default/namespace:oip/j:01KEEH8F7Z9XXXXXXXXXXXXXXX"

# 3. 查看 Job 元数据
redis-cli HGETALL "pool:default/namespace:oip/m:01KEEH8F7Z9XXXXXXXXXXXXXXX"
# 输出示例：
# 1) "tries"
# 2) "3"
# 3) "ttl"
# 4) "3600"

# 4. 查看 Working 状态（消费中）
redis-cli KEYS "pool:default/namespace:oip/w:*"
redis-cli TTL "pool:default/namespace:oip/w:test_queue/01KEEH8F7Z9XXXXXXXXXXXXXXX"

# 5. 查看延迟队列
redis-cli ZRANGE "pool:default/namespace:oip/t:test_queue" 0 -1 WITHSCORES
```

**状态流转示意：**
```
发布 → pool:default/namespace:oip/q:test_queue (Ready Queue)
消费 → pool:default/namespace:oip/w:test_queue/{job_id} (Working, TTR 倒计时)
ACK  → 删除所有相关键
超时 → 回到 Ready Queue (tries--)
```

---

### 2.3 死信队列（Dead Letter）

#### 进入死信队列的条件

**核心规则：重试次数耗尽**

当 job 满足以下条件时进入死信队列：
1. **TTR 超时或显式 NACK**：Job 未在 `ttr` 时间内 ACK，或 Consumer 主动 NACK
2. **Tries 递减**：每次重新入队，`tries--`
3. **Tries 耗尽**：当 `tries = 0` 时，job 进入死信队列

**条件分析：**
```
进入死信队列 ⇔ tries = 0
```

- ❌ **错误理解**：`ttr` 超时就进死信队列
- ✅ **正确理解**：`ttr` 超时只是触发 requeue，`tries--`；只有 `tries` 用完才进死信队列

**官方文档原文解读：**
> When a job failed to ACK within ttr and no more tries are available, the job will be put into a sink called "dead letter".

翻译：当 job 在 `ttr` 时间内未 ACK **且** `tries` 用完时，进入死信队列。

#### 原型定义
```
GET    /api/:namespace/:queue/deadletter             # 查看死信队列
PUT    /api/:namespace/:queue/deadletter?limit=&ttl= # 复活 job
DELETE /api/:namespace/:queue/deadletter?limit=      # 删除死信 job
```

#### 实战场景

**场景设置：**
```bash
# 发布一个 tries=2 的 job（方便快速测试）
curl -XPUT \
  -H "X-Token:01KEED5FWJB9GT21S0GVQ4SXXW" \
  -d '{"order_id": "DEAD001"}' \
  "http://127.0.0.1:7777/api/oip/test_queue?delay=0&ttl=3600&tries=2"
```

**测试流程：**

**第 1 次消费（tries=1）**
```bash
# 消费 job，ttr=5 秒
curl -H "X-Token:01KEED5FWJB9GT21S0GVQ4SXXW" \
  "http://127.0.0.1:7777/api/oip/test_queue?ttr=5&timeout=2"

# 响应：返回 job_id，假设为 01KEEH9XXXX

# 等待 5 秒不 ACK，让 ttr 超时
sleep 6
```

**第 2 次消费（tries=0）**
```bash
# 再次消费（此时 tries 已递减为 1）
curl -H "X-Token:01KEED5FWJB9GT21S0GVQ4SXXW" \
  "http://127.0.0.1:7777/api/oip/test_queue?ttr=5&timeout=2"

# 再次等待 5 秒不 ACK
sleep 6

# 此时 tries=0，job 进入死信队列
```

**1. 查看死信队列**
```bash
curl -H "X-Token:01KEED5FWJB9GT21S0GVQ4SXXW" \
  "http://127.0.0.1:7777/api/oip/test_queue/deadletter"
```

**响应：**
```json
{
  "namespace": "oip",
  "queue": "test_queue",
  "deadletter_size": 1,
  "deadletter_head": "01KEEH9XXXXXXXXXXXX"
}
```

**2. 复活死信 job**
```bash
# 复活 1 个 job，设置新的 ttl=86400, tries=1
curl -XPUT \
  -H "X-Token:01KEED5FWJB9GT21S0GVQ4SXXW" \
  "http://127.0.0.1:7777/api/oip/test_queue/deadletter?limit=1&ttl=86400"
```

**响应：**
```json
{
  "msg": "respawned",
  "count": 1
}
```

**注意事项：**
- 复活后的 job 默认配置：`ttl=86400&tries=1&delay=0`
- 可通过 URL 参数自定义 `ttl` 和 `tries`
- 死信队列中的 job **无 TTL**（永不过期），需手动处理

**3. 删除死信 job**
```bash
# 删除 1 个死信 job
curl -XDELETE \
  -H "X-Token:01KEED5FWJB9GT21S0GVQ4SXXW" \
  "http://127.0.0.1:7777/api/oip/test_queue/deadletter?limit=1"
```

**响应：**
```
HTTP/1.1 204 No Content
```

#### Redis 数据检查

```bash
# 1. 查看死信队列
redis-cli LLEN "pool:default/namespace:oip/d:test_queue"
redis-cli LRANGE "pool:default/namespace:oip/d:test_queue" 0 -1

# 示例输出：
# 1) "01KEEH9XXXXXXXXXXXX"

# 2. 查看死信 job 数据
redis-cli GET "pool:default/namespace:oip/j:01KEEH9XXXXXXXXXXXX"

# 3. 查看死信 job 元数据
redis-cli HGETALL "pool:default/namespace:oip/m:01KEEH9XXXXXXXXXXXX"
# 输出：
# 1) "tries"
# 2) "0"       # tries 已耗尽
# 3) "ttl"
# 4) "0"       # 死信队列中 TTL 移除
```

**死信队列特性总结：**
- ✅ Job 在死信队列中**无 TTL**（永不过期）
- ✅ 可通过 API 查看、复活、删除
- ✅ 复活时可重新设置 `ttl` 和 `tries`
- ⚠️ 如果 job 已被删除（TTL 过期或手动删除），复活操作无效

---

## 三、完整状态流转图

```
┌────────────────────────────────────────────────────────────────┐
│                     Job 生命周期                                │
└────────────────────────────────────────────────────────────────┘

   Publish (tries=3, ttl=3600, ttr=30)
          │
          ▼
   ┌──────────────────┐
   │  Ready Queue     │  ← Redis: pool:default/namespace:oip/q:{queue}
   │  (等待消费)      │
   └──────────────────┘
          │
          │ Consume (timeout=3, ttr=30)
          ▼
   ┌──────────────────┐
   │  Working         │  ← Redis: pool:default/namespace:oip/w:{queue}/{job_id}
   │  tries=2         │     TTL=30 (ttr 倒计时)
   │  TTR 倒计时      │
   └──────────────────┘
          │
    ┌─────┴─────┐
    │           │
  ACK       TTR 超时 / NACK
    │           │
    ▼           ▼ (tries--)
 ┌─────┐   ┌──────────────────┐
 │删除 │   │  Requeue         │
 └─────┘   │  tries=1         │
           └──────────────────┘
                  │
                  │ 再次消费 (ttr=30)
                  ▼
           ┌──────────────────┐
           │  Working         │
           │  tries=0         │
           └──────────────────┘
                  │
             TTR 超时 / NACK
                  │ (tries=0)
                  ▼
           ┌──────────────────┐
           │  Dead Letter     │  ← Redis: pool:default/namespace:oip/d:{queue}
           │  (死信队列)      │     TTL=0 (永不过期)
           └──────────────────┘
                  │
           ┌──────┴──────┐
           │             │
       Respawn       Delete
           │             │
           ▼             ▼
     回到 Ready      永久删除
```

---

## 四、项目实践总结

### 4.1 配置管理策略
- **开发环境**：配置文件（便于团队协作，避免环境变量配置负担）
- **生产环境**：配置中心（K8s ConfigMap/Secret）或 Vault

### 4.2 Lmstfy 使用建议

**参数设置：**
- `ttr`：任务平均处理时间的 **1.5-2 倍**（留出容错空间）
- `tries`：根据业务容忍度设置（建议 **3-5 次**）
- `ttl`：根据业务需求设置，0 表示永不过期

**监控告警：**
- 定期检查死信队列大小（`/deadletter` API）
- 设置死信队列告警阈值（如 > 100）
- 定期清理或复活死信 job

**最佳实践：**
```go
// Consumer 配置建议
&consumer.Config{
    QueueName:    "order_diagnose_callback",
    Timeout:      3,   // long polling 3 秒
    TTR:          30,  // 任务处理超时 30 秒
    PollInterval: 100 * time.Millisecond,
}

// Job 发布建议
lmstfy.Publish(queue, job, &PublishOptions{
    TTL:   3600,  // 1 小时内有效
    Delay: 0,     // 立即可消费
    Tries: 3,     // 最多重试 3 次
})
```

### 4.3 待优化项
- [ ] 敏感配置迁移到配置中心
- [ ] 添加 Lmstfy 健康检查（`/health` endpoint）
- [ ] 死信队列监控告警
- [ ] Job 处理时长监控（辅助 `ttr` 参数调优）

---

## 五、参考资料

- [Lmstfy GitHub](https://github.com/bitleak/lmstfy)
- [Lmstfy API 文档](https://github.com/bitleak/lmstfy/blob/master/doc/API.md)
- [Lmstfy 管理 API](https://github.com/bitleak/lmstfy/blob/master/doc/administration.en.md)
- [Redis 数据结构](https://redis.io/docs/data-types/)

---

**文档更新时间**：2026-01-08
**作者**：cooperswang
**版本**：v1.0
