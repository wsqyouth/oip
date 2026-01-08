# OIP Backend - 产品需求文档

> 跨境订单智能诊断平台 (Order Intelligence Platform)
> 版本: v3.0
> 最后更新: 2026-01-06

## 产品定位

**轻量级跨境订单智能诊断 SaaS 平台**

为中小跨境卖家提供标准化订单接入、智能诊断（物流费率/异常检测）和数据可视化能力。

**MVP 目标**：跑通 "接入 → 诊断 → 结果展示" 完整闭环。

**明确边界（MVP 不涉及）**：
- ❌ 真实物流下单执行
- ❌ 库存管理
- ❌ 真实物流 API 对接（使用 Mock）
- ❌ API Token 认证（Phase 2）

---

## 核心功能

### 1. 账号管理
- 创建账号（name + email）
- 查询账号信息

### 2. 订单诊断
- **订单接入**: 接收商家订单（包含发货地址、收货地址、包裹信息）
- **智能等待**: Smart Wait 机制，Hold 连接 10s 等待诊断结果
- **异步诊断**:
  - 物流费率计算（Mock 多家承运商报价）
  - 异常检测（高价值、大件、SKU 缺失）
- **结果返回**:
  - 10s 内完成 → 返回完整诊断结果
  - 超时 → 返回 Processing 状态，支持轮询查询

### 3. 状态管理
- `DIAGNOSING`: 诊断中
- `DIAGNOSED`: 诊断完成
- `FAILED`: 诊断失败

---

## 系统架构

### 架构模式

**Shared Kernel（共享内核）+ Monorepo**

```
oip/backend/
├── common/          # 共享内核（Entity, DAO, Model）
├── dpmain/          # 同步 API 服务（HTTP + Smart Wait）
└── dpsync/          # 异步 Worker 服务（消费队列 + 诊断逻辑）
```

### 技术栈

| 层级 | 技术 | 说明 |
|------|------|------|
| **语言** | Go 1.21+ | - |
| **Web 框架** | Gin | HTTP API |
| **ORM** | GORM | 数据访问 |
| **DI** | Wire | 依赖注入 |
| **消息队列** | Lmstfy | 异步任务 |
| **数据库** | MySQL 8.0 | 订单数据存储 |
| **缓存/通知** | Redis 7.0 | Pub/Sub 通知 |
| **日志** | Zap | 结构化日志 |

### 核心流程

```
商家系统 → POST /api/v1/orders?wait=10
              ↓
          dpmain (API 服务)
              ├─ 验证账号
              ├─ 地址校验
              ├─ 订单落库 (status=DIAGNOSING)
              ├─ 订阅 Redis channel
              ├─ 推送 Lmstfy 队列
              └─ Smart Wait (10s)
                  ↓
            10s 内收到结果 → 200 OK + 诊断结果
            超时 → 3001 Processing + poll_url
              ↓
          dpsync (Worker 服务)
              ├─ 消费 Lmstfy 队列
              ├─ 执行诊断逻辑
              │   ├─ 费率计算 (Mock)
              │   └─ 异常检测 (规则引擎)
              ├─ 更新数据库 (status=DIAGNOSED)
              └─ 发布 Redis 通知
```

---

## 数据库设计

### Accounts 表

```sql
CREATE TABLE `accounts` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `name` varchar(255) NOT NULL COMMENT '账号名称',
  `email` varchar(255) UNIQUE NOT NULL COMMENT '邮箱',
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_email` (`email`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='账号表';
```

### Orders 表（包含诊断结果）

```sql
CREATE TABLE `orders` (
  -- 基础字段
  `id` varchar(64) NOT NULL COMMENT '订单ID (UUID)',
  `account_id` bigint unsigned NOT NULL COMMENT '账号ID',
  `merchant_order_no` varchar(128) NOT NULL COMMENT '商家订单号',

  -- 订单数据
  `raw_data` json NOT NULL COMMENT '原始订单数据（shipment信息）',

  -- 诊断状态与结果
  `status` varchar(16) NOT NULL DEFAULT 'DIAGNOSING'
    COMMENT '订单诊断状态: DIAGNOSING, DIAGNOSED, FAILED',
  `diagnose_result` json DEFAULT NULL
    COMMENT '诊断结果（DiagnosisResultData）',
  `error_message` text
    COMMENT '整体失败时的错误信息',

  -- 时间戳
  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,

  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_account_merchant` (`account_id`, `merchant_order_no`),
  INDEX `idx_account_status` (`account_id`, `status`),
  INDEX `idx_created_at` (`created_at`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='订单表';
```

### 诊断结果结构（JSON）

```json
{
  "items": [
    {
      "type": "shipping",
      "status": "SUCCESS",
      "data_json": {
        "recommended_code": "FEDEX_GROUND",
        "rates": [
          {
            "carrier": "FedEx",
            "service": "Ground",
            "total_fee": 12.50,
            "transit_days": 3,
            "tags": ["CHEAPEST"]
          }
        ]
      },
      "error": null
    },
    {
      "type": "anomaly",
      "status": "SUCCESS",
      "data_json": {
        "has_risk": true,
        "issues": [
          {
            "type": "HIGH_VALUE",
            "level": "WARNING",
            "message": "Order value exceeds $500"
          }
        ]
      },
      "error": null
    }
  ]
}
```

---

## API 设计

### 统一响应格式

参考 AfterShip API 标准，采用嵌套 `meta` 结构：

**成功响应：**
```json
{
  "meta": {
    "code": 200,
    "message": "OK"
  },
  "data": {
    ...
  }
}
```

**错误响应（带详情）：**
```json
{
  "meta": {
    "code": 400,
    "message": "Bad Request",
    "details": [
      {
        "path": "email",
        "info": "email is required"
      },
      {
        "path": "shipment.ship_from.postal_code",
        "info": "postal_code must be valid"
      }
    ]
  }
}
```

**响应结构说明：**
- `meta`: 元数据对象（必须）
  - `code`: 业务状态码（必须）- 200成功, 400客户端错误, 500服务器错误, 3001处理中
  - `message`: 响应消息（必须）
  - `details`: 错误详情数组（可选）- 仅在参数验证失败时返回
- `data`: 业务数据对象（可选）- 成功时包含具体数据

### 核心接口

#### 1. 创建订单（Smart Wait）

```http
POST /api/v1/orders?wait=10
Content-Type: application/json

{
  "account_id": 1,
  "merchant_order_no": "ORD-20260106-001",
  "shipment": {
    "ship_from": { ... },
    "ship_to": { ... },
    "parcels": [ ... ]
  }
}
```

**响应 A：10s 内完成（诊断成功）**
```json
{
  "meta": {
    "code": 200,
    "message": "OK"
  },
  "data": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "account_id": 1,
    "merchant_order_no": "ORD-20260106-001",
    "status": "COMPLETED",
    "diagnosis": {
      "items": [
        {
          "type": "shipping",
          "status": "SUCCESS",
          "data_json": {
            "recommended_code": "FEDEX_GROUND",
            "rates": [...]
          }
        },
        {
          "type": "anomaly",
          "status": "SUCCESS",
          "data_json": {
            "has_risk": false,
            "issues": []
          }
        }
      ]
    },
    "created_at": "2024-01-01T00:00:00Z",
    "updated_at": "2024-01-01T00:00:10Z"
  }
}
```

**响应 B：超时（诊断进行中）**
```json
{
  "meta": {
    "code": 3001,
    "message": "Order is being diagnosed, please poll for results"
  },
  "data": {
    "order_id": "550e8400-e29b-41d4-a716-446655440000",
    "poll_url": "/api/v1/orders/550e8400-e29b-41d4-a716-446655440000"
  }
}
```

**说明：**
- `code=3001` 表示订单诊断进行中，需要客户端通过 `poll_url` 轮询获取最终结果
- `poll_url` 包含在 `data` 中，方便客户端直接使用

#### 2. 查询订单

```http
GET /api/v1/orders/{id}
```

#### 3. 创建账号

```http
POST /api/v1/accounts
Content-Type: application/json

{
  "name": "John's Store",
  "email": "john@example.com"
}
```

---

## 关键设计决策

### 1. 单表设计
- **不使用独立的 Diagnoses 表**
- 诊断结果直接存储在 `orders.diagnose_result` (JSON)
- 简化查询逻辑，避免 JOIN

### 2. Smart Wait 机制
- dpmain Hold 连接 10s 等待诊断结果
- 使用 **Redis Pub/Sub**（独立 channel: `diagnosis:result:{order_id}`）
- 超时降级返回 `3001 Processing`，前端轮询

### 3. 异步架构
- **计算密集型**: order_diagnose_channel（费率计算 + 异常检测）
- **IO 密集型**: order_notify_channel（推送通知，Phase 2）

### 4. 扩展性设计
- 诊断结果使用 `items` 数组 + `type` 字段
- 新增诊断类型时，只需追加数组元素，不修改表结构
- 支持的类型：`shipping`, `anomaly`, `risk`(Phase 2), `compliance`(Phase 2)

---

## MVP 实施计划

### Week 1: 基础设施 + Account + Order 接入
- [ ] 初始化 Monorepo（go.work + 三模块）
- [ ] 搭建 MySQL + Redis 环境（Docker Compose）
- [ ] 实现 common/entity 和 common/dao
- [ ] 实现 Account API（创建/查询）
- [ ] 实现 Order 创建 API（不含 Smart Wait）

### Week 2: Smart Wait + Redis 通知
- [ ] Redis Pub/Sub 封装
- [ ] 修改 Order 创建 API，增加 Smart Wait 逻辑
- [ ] 单元测试：验证 200 vs 3001
- [ ] 压力测试：验证长连接稳定性

### Week 3: dpsync Worker + Mock 诊断
- [ ] 基于 DPSync 框架实现 Worker
- [ ] 实现 CompositeHandler
- [ ] 实现 ShippingCalculator（Mock 费率）
- [ ] 实现 AnomalyChecker（规则引擎）
- [ ] 端到端测试

### Week 4: 前端 Dashboard（简易版）
- [ ] React + Ant Design 搭建
- [ ] 订单列表页
- [ ] 订单详情页
- [ ] 数据统计

---

## Story 拆分索引

> 详细任务拆分见 `stories/` 目录

- [ ] Story-001: 基础设施搭建（MySQL + Redis + Lmstfy）
- [ ] Story-002: Account API 实现
- [ ] Story-003: Order 接入 API（不含 Smart Wait）
- [ ] Story-004: Smart Wait 机制实现
- [ ] Story-005: dpsync Worker 框架搭建
- [ ] Story-006: ShippingCalculator Mock 实现
- [ ] Story-007: AnomalyChecker 规则引擎
- [ ] Story-008: 端到端集成测试

---

## Phase 2 扩展计划

| 功能 | 说明 | 优先级 |
|------|------|--------|
| **Webhook 推送** | 诊断完成后推送到商家系统 | P1 |
| **真实物流 API** | 对接 FedEx/UPS/EasyPost | P1 |
| **API Token 认证** | Account 权限隔离 | P2 |
| **高级地址清洗** | 接入 Google Maps API | P2 |
| **风险评分诊断** | 新增 `risk` 诊断类型 | P3 |

---

## 术语表

| 术语 | 说明 |
|------|------|
| **Account** | 账号（对应一个商家） |
| **Order** | 订单 |
| **Diagnosis** | 诊断 |
| **Smart Wait** | 智能等待（Hold 连接等结果） |
| **CompositeHandler** | 组合处理器（执行多个诊断逻辑） |
| **dpmain** | 同步 API 服务 |
| **dpsync** | 异步 Worker 服务 |
| **Shared Kernel** | 共享内核（DDD 概念） |
