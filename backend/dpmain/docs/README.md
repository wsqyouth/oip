# DPMain 架构设计文档

## 文档概述

本文档是 DPMain（Order Intelligent Platform）项目的完整架构设计文档，涵盖业务背景、DDD 理论、架构设计、技术决策、扩展性设计和面试准备等内容。

**文档特点：**
- ✅ 解释"为什么"这样设计，而不仅仅是"是什么"
- ✅ 包含 DDD 理论与实战应用
- ✅ 讨论架构权衡与实用主义
- ✅ 使用 Mermaid 绘制架构图
- ✅ 包含 k6 性能压测指南
- ✅ 提供面试准备（STAR 方法）

## 文档结构

### [第一部分：业务背景与架构思路](./01_business_background_and_approach.md)
**阅读时间：15 分钟**

- 1.1 问题域：国际物流诊断系统
- 1.2 核心业务流程
- 1.3 核心业务挑战（Smart Wait、数据一致性、错误处理）
- 2.1 从问题到方案的演进路径（单体架构、分层架构）
- 2.2 核心设计原则（依赖倒置、单一职责、实用主义）
- 2.3 设计演进的关键里程碑

**适合人群：** 想要了解项目背景和整体设计思路的读者

---

### [第二部分：DDD 理论与应用](./02_ddd_theory_and_application.md)
**阅读时间：20 分钟**

- 3.1 为什么需要 DDD？（贫血模型 vs 富领域模型）
- 3.2 DDD 核心概念（实体、值对象、聚合根、领域服务、仓储）
- 3.3 领域建模实战（识别核心概念、区分实体与值对象、设计聚合边界）
- 3.4 DDD 的实用主义权衡（贫血模型 vs 富领域模型、严格分层 vs 实用主义、领域事件 vs 直接调用）

**适合人群：** 想要深入理解 DDD 理论和实践的读者

---

### [第三部分：核心架构设计](./03_core_architecture_design.md)
**阅读时间：25 分钟**

- 4.1 完整分层架构（API → Service → Module → Repository → Infrastructure）
- 4.2 为什么需要 Module 层？（Service vs Module 的区别）
- 4.3 Repository 实现的位置权衡（domains/repo vs infra）
- 4.4 DiagnosisModule 为什么直接依赖基础设施？
- 4.5 依赖注入的设计（Wire 的优势）
- 4.6 单进程多 Goroutine 架构（HTTP Server + Consumer）

**适合人群：** 想要了解分层架构设计和依赖注入的读者

---

### [第四部分：关键技术决策](./04_key_technical_decisions.md)
**阅读时间：25 分钟**

- 5.1 Smart Wait 机制（Redis Pub/Sub + 超时降级）
- 5.2 错误处理策略（可恢复错误 vs 致命错误）
- 5.3 消息自包含设计（为什么 Shipment 要传递完整数据）
- 5.4 订单状态机设计（DIAGNOSING → PENDING → SHIPPED）
- 5.5 性能优化策略（连接池、索引、避免 N+1 查询）

**适合人群：** 想要了解关键技术实现细节的读者

---

### [第五部分：扩展性设计与面试准备](./05_extensibility_and_interview.md)
**阅读时间：30 分钟**

- 6.1 数据库切换（MySQL → PostgreSQL）
- 6.2 添加监控和链路追踪（Prometheus、Jaeger）
- 6.3 分布式部署（多实例、负载均衡）
- 7.1 使用 STAR 方法讲述项目亮点
  - 亮点 1：Smart Wait 机制
  - 亮点 2：DDD 分层架构与实用主义权衡
  - 亮点 3：单进程多 Goroutine 架构
- 7.2 性能压测（k6）
  - 场景 1：创建订单（无 Smart Wait）
  - 场景 2：创建订单（Smart Wait）
  - 场景 3：查询订单列表

**适合人群：** 想要了解系统扩展性设计和准备面试的读者

---

### [第六部分：总结与反思](./06_summary_and_reflection.md)
**阅读时间：20 分钟**

- 8.1 核心设计思想总结
  - 实用主义的 DDD
  - 事件驱动的异步架构
  - 单进程多 Goroutine 的简洁架构
  - 显式错误处理
- 8.2 项目中的经验教训
  - 架构演进需要迭代
  - 错误处理需要显式设计
  - 测试覆盖率需要持续投入
  - 监控和日志需要提前规划
- 8.3 未来的优化方向
  - 引入分布式事务（Saga 模式）
  - 引入 CQRS（读写分离）
  - 引入限流和熔断
  - 引入多租户支持
- 8.4 给后来者的建议
  - 先理解业务，再设计架构
  - 分层架构不是银弹
  - 测试是架构的一部分
  - 架构需要迭代演进
  - 文档是架构的延续
- 9. 附录
  - 目录结构
  - 技术栈清单
  - 关键指标
  - 常见问题 FAQ

**适合人群：** 想要全面总结和反思项目的读者

---

## 阅读建议

### 快速入门（30 分钟）
如果你想快速了解项目，建议阅读：
1. [第一部分](./01_business_background_and_approach.md) - 1.1, 1.2, 2.1
2. [第三部分](./03_core_architecture_design.md) - 4.1
3. [第六部分](./06_summary_and_reflection.md) - 8.1

### 深入理解（2 小时）
如果你想深入理解架构设计，建议完整阅读：
1. [第一部分](./01_business_background_and_approach.md) - 业务背景与架构思路
2. [第二部分](./02_ddd_theory_and_application.md) - DDD 理论与应用
3. [第三部分](./03_core_architecture_design.md) - 核心架构设计
4. [第四部分](./04_key_technical_decisions.md) - 关键技术决策

### 面试准备（1 小时）
如果你是为了面试准备，建议重点阅读：
1. [第五部分](./05_extensibility_and_interview.md) - 7.1（STAR 方法讲述项目亮点）
2. [第六部分](./06_summary_and_reflection.md) - 8.2（项目中的经验教训）
3. [第六部分](./06_summary_and_reflection.md) - 9.4（常见问题 FAQ）

### 实战演练（2 小时）
如果你想动手实践，建议：
1. 阅读 [第五部分](./05_extensibility_and_interview.md) - 7.2（性能压测）
2. 运行 k6 压测脚本
3. 阅读 [第六部分](./06_summary_and_reflection.md) - 9.1（目录结构）
4. 查看实际代码实现

---

## 核心亮点

### 🎯 Smart Wait 机制
在异步架构的基础上提供"伪同步"的用户体验，通过 Redis Pub/Sub + 超时降级，让 85% 的订单在 5 秒内获得诊断结果。

### 🏗️ 实用主义的 DDD
在保留 DDD 核心价值的同时，根据项目规模和团队能力进行权衡，避免过度设计。

### ⚡ 单进程多 Goroutine
HTTP Server 和 Callback Consumer 在同一进程中，共享连接池，简化部署，提高资源利用率。

### 🔍 显式错误处理
区分可恢复错误和致命错误，所有错误都必须被处理（返回或记录日志），便于监控告警。

---

## 技术栈

- **语言**：Go 1.21
- **Web 框架**：Gin
- **ORM**：GORM
- **数据库**：MySQL 8.0
- **缓存**：Redis 6.0
- **消息队列**：Lmstfy
- **依赖注入**：Wire
- **监控**：Prometheus + Grafana
- **链路追踪**：Jaeger + OpenTelemetry
- **压测**：k6

---

## 贡献与反馈

如果你发现文档中的错误或有改进建议，欢迎：
1. 提交 Issue
2. 提交 Pull Request
3. 联系项目维护者

---

## 版本历史

| 版本 | 日期 | 作者 | 变更内容 |
|------|------|------|---------|
| v1.0 | 2025-12-30 | Claude Sonnet 4.5 | 初始版本，完整架构设计文档 |

---

## 许可证

本文档采用 [CC BY-SA 4.0](https://creativecommons.org/licenses/by-sa/4.0/) 许可证。

---

**Happy Learning! 🚀**
