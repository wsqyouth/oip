# Task-002: 生成 Swagger 文档 PRD

> 创建日期: 2026-01-07
> 负责人: cooperswang
> 状态: 📝 待开发
> 项目: OIP Backend - 跨境订单智能诊断平台

---

## 🎯 目标

为 OIP Backend API 生成完整的 Swagger 文档，提供类似 AfterShip API 文档的清晰、专业的 API 接口说明，方便前端开发和第三方集成。

### 核心目标
1. 覆盖所有现有 API 接口（Accounts 和 Orders）
2. 提供完整的字段定义和类型说明
3. 包含真实可运行的请求/响应示例
4. 支持在线交互测试（Try it out）
5. 可导出 OpenAPI 规范文档

---

## 📖 背景

### 当前状态
- ✅ API 接口已实现（5个端点）
- ✅ Request/Response 模型已定义
- ✅ 统一响应格式已实现（ginx.Response）
- ❌ 缺少 API 文档
- ❌ 前端和第三方无法了解接口规范

### 期望状态
- ✅ 访问 `http://localhost:8080/swagger/index.html` 可查看完整 API 文档
- ✅ 每个接口有清晰的字段说明和示例
- ✅ 支持在线测试 API
- ✅ 可导出 OpenAPI 文档供工具使用

### 参考标准
参考 AfterShip API 文档风格：
- URL: https://www.aftership.com/docs/shipping/
- 特点：清晰分组、详细字段说明、完整示例、交互测试

---

## ✅ DoD（Definition of Done）

### 1. 功能覆盖

**必须覆盖的接口：**
- [ ] `GET /health` - 健康检查
- [ ] `POST /api/v1/accounts` - 创建账号
- [ ] `GET /api/v1/accounts/:id` - 获取账号详情
- [ ] `POST /api/v1/orders` - 创建订单（核心接口）
- [ ] `GET /api/v1/orders/:id` - 获取订单详情（含诊断结果）

### 2. 文档质量标准

**2.1 每个接口必须包含：**
- [ ] 清晰的功能描述（Summary + Description）
- [ ] 完整的 HTTP Method 和 Path
- [ ] Tags 分组（accounts / orders）
- [ ] 请求示例（Request Example）
- [ ] 响应示例（Response Example）- 至少包含成功场景

**2.2 每个字段必须标注：**
- [ ] 数据类型（string, integer, number, boolean, object, array）
- [ ] 是否必填（required 标识）
- [ ] 字段说明（清晰的 description）
- [ ] 特殊约束：
  - [ ] email 格式验证（`binding:"email"`）
  - [ ] 枚举值（如 `status: PENDING, DIAGNOSING, COMPLETED, FAILED`）
  - [ ] 数值范围（如 `weight.value > 0`）
  - [ ] 字符串长度限制

**2.3 Schema 对象独立定义：**

**Request Models:**
- [ ] `CreateAccountRequest`
- [ ] `CreateOrderRequest`
- [ ] `Shipment`
- [ ] `Address`
- [ ] `Parcel`
- [ ] `Weight`
- [ ] `Dimension`
- [ ] `Item`
- [ ] `Money`

**Response Models:**
- [ ] `Response` - 统一响应格式
- [ ] `AccountResponse`
- [ ] `OrderResponse`
- [ ] `DiagnosisResult`
- [ ] `DiagnosisItem`

**2.4 响应状态码完整定义：**
- [ ] 200 OK - 成功（包含具体返回数据）
- [ ] 400 Bad Request - 参数错误（包含错误消息）
- [ ] 404 Not Found - 资源不存在
- [ ] 500 Internal Server Error - 服务器错误
- [ ] 特殊：订单创建接口需要说明 `code: 3001` 的 Processing 状态（Smart Wait 超时场景）

### 3. 技术实现

**3.1 工具和依赖：**
- [ ] 安装 `swaggo/swag` CLI 工具
- [ ] 添加依赖到 `dpmain/go.mod`：
  - `github.com/swaggo/swag`
  - `github.com/swaggo/gin-swagger`
  - `github.com/swaggo/files`

**3.2 API 总体信息配置：**
- [ ] 在 `cmd/apiserver/main.go` 添加 API 元信息注释
- [ ] 配置项：
  - Title: `OIP Backend API`
  - Version: `1.0`
  - Description: `跨境订单智能诊断平台后端 API`
  - Host: `localhost:8080`
  - BasePath: `/api/v1`
  - Security: `ApiKeyAuth` (header: `api-key`)

**3.3 Handler 注释（Swagger Annotations）：**
- [ ] 每个 handler 方法添加完整的 Swagger 注释
- [ ] 注释格式符合 swaggo 规范
- [ ] 包含：`@Summary`, `@Description`, `@Tags`, `@Accept`, `@Produce`, `@Param`, `@Success`, `@Failure`, `@Security`, `@Router`

**3.4 路由配置：**
- [ ] 在 `routers/router.go` 中添加 Swagger UI 路由
- [ ] 路径：`/swagger/*any`
- [ ] Handler: `ginSwagger.WrapHandler`

**3.5 构建配置：**
- [ ] 在 `Makefile` 添加 `swagger` 命令
- [ ] 生成目标目录：`docs/`
- [ ] 入口文件：`cmd/apiserver/main.go`

### 4. 认证配置

- [ ] 定义 API Key 认证方式（header: `api-key`）
- [ ] 每个业务接口添加 `@Security ApiKeyAuth` 标记
- [ ] 在文档说明中注明：当前版本认证暂未启用，仅作为占位

### 5. 完整示例

**5.1 请求示例：**
- [ ] Account 创建：包含 name 和 email
- [ ] Order 创建：包含完整的 shipment 信息（ship_from, ship_to, parcels）
- [ ] 使用真实的示例数据（美国地址、常见商品等）
- [ ] 在 README 中提供 curl 命令示例

**5.2 响应示例：**
- [ ] 成功响应：完整的返回数据结构
- [ ] 错误响应：常见错误场景（400/404/500）
- [ ] 订单创建接口：同时展示 200 成功和 3001 Processing 两种场景

### 6. 验收测试

**6.1 文档生成：**
- [ ] 运行 `make swagger` 成功生成文档
- [ ] 无编译错误和警告
- [ ] 生成文件存在：
  - `docs/swagger.json`
  - `docs/swagger.yaml`
  - `docs/docs.go`

**6.2 本地访问：**
- [ ] 启动服务：`make run-dpmain`
- [ ] 访问 `http://localhost:8080/swagger/index.html` 可正常打开
- [ ] Swagger UI 正确渲染所有 5 个接口
- [ ] 接口按 Tags 正确分组（accounts / orders）

**6.3 文档交互测试：**
- [ ] 点击每个接口可展开查看详情
- [ ] 所有字段类型、required 标识正确显示
- [ ] Example Value 可正常显示请求体示例
- [ ] Try it out 功能可用（能输入参数并发送测试请求）
- [ ] 响应状态码说明完整
- [ ] Models 部分可查看所有 Schema 定义

**6.4 文档导出：**
- [ ] 可通过 Swagger UI 下载 `swagger.json`
- [ ] 可通过 Swagger UI 下载 `swagger.yaml`
- [ ] 文档格式符合 OpenAPI 3.0 规范
- [ ] 可导入 Postman 等工具使用

### 7. 项目文档更新

- [ ] 更新 `dpmain/README.md`，添加 Swagger 使用说明
- [ ] 包含：如何生成文档、如何访问文档、API 概览
- [ ] 更新根目录 `README.md`（如有）

### 8. 代码质量

- [ ] 通过 `gofmt -w .` 格式化
- [ ] 通过 `go vet ./...` 静态检查
- [ ] 通过 `go build` 编译成功
- [ ] Swagger 注释不影响代码可读性
- [ ] 无硬编码配置
- [ ] 不引入破坏性变更

---

## 📋 任务拆分

### 阶段 1: 环境准备（预计 30 分钟）

#### Task 1.1: 安装 Swagger 工具
- [x] 安装 swag CLI 工具
  ```bash
  go install github.com/swaggo/swag/cmd/swag@latest
  ```
- [x] 验证安装成功：`swag --version`
- [x] 确认 `$GOPATH/bin` 在 PATH 中

#### Task 1.2: 添加 Go 依赖
- [x] 在 `dpmain/` 目录执行：
  ```bash
  go get -u github.com/swaggo/swag
  go get -u github.com/swaggo/gin-swagger
  go get -u github.com/swaggo/files
  ```
- [x] 运行 `go mod tidy` 清理依赖
- [ ] 验证 `go.mod` 已添加依赖

#### Task 1.3: 配置 Makefile
- [ ] 在根目录 `Makefile` 添加 swagger 命令：
  ```makefile
  .PHONY: swagger
  swagger:
      @echo "Generating swagger docs..."
      cd dpmain && swag init -g cmd/apiserver/main.go -o docs --parseDependency --parseInternal
      @echo "Swagger docs generated at dpmain/docs/"
  ```
- [ ] 测试运行：`make swagger`（预期会有警告，因为还没添加注释）

---

### 阶段 2: 添加 API 总体配置（预计 20 分钟）

#### Task 2.1: 配置 main.go 注释
- [ ] 编辑 `dpmain/cmd/apiserver/main.go`
- [ ] 在 `main()` 函数上方添加：
  ```go
  // @title           OIP Backend API
  // @version         1.0
  // @description     跨境订单智能诊断平台后端 API，提供订单接入和智能诊断服务
  // @termsOfService  http://swagger.io/terms/
  
  // @contact.name   API Support
  // @contact.email  support@oip.example.com
  
  // @license.name  Apache 2.0
  // @license.url   http://www.apache.org/licenses/LICENSE-2.0.html
  
  // @host      localhost:8080
  // @BasePath  /api/v1
  
  // @securityDefinitions.apikey ApiKeyAuth
  // @in header
  // @name api-key
  // @description API Key 用于接口认证（当前版本暂未启用，保留占位）
  ```

#### Task 2.2: 配置 Swagger 路由
- [ ] 编辑 `dpmain/internal/app/server/routers/router.go`
- [ ] 添加 import：
  ```go
  import (
      swaggerFiles "github.com/swaggo/files"
      ginSwagger "github.com/swaggo/gin-swagger"
      _ "oip/dpmain/docs"  // 导入生成的 docs
  )
  ```
- [ ] 在 `SetupRoutes()` 中添加路由：
  ```go
  // Swagger 文档路由
  r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
  ```

---

### 阶段 3: 完善 Request/Response 模型注释（预计 40 分钟）

#### Task 3.1: Account Request/Response 注释

**文件：`dpmain/internal/app/domains/apimodel/request/create_account_request.go`**
- [ ] 添加结构体注释：
  ```go
  // CreateAccountRequest 创建账号请求
  type CreateAccountRequest struct {
      Name  string `json:"name" binding:"required" example:"John Doe"`          // 账号名称
      Email string `json:"email" binding:"required,email" example:"john@example.com"` // 账号邮箱
  }
  ```

**文件：`dpmain/internal/app/domains/apimodel/response/account_response.go`**
- [ ] 添加结构体注释：
  ```go
  // AccountResponse 账号响应
  type AccountResponse struct {
      ID        int64     `json:"id" example:"1"`                                    // 账号ID
      Name      string    `json:"name" example:"John Doe"`                           // 账号名称
      Email     string    `json:"email" example:"john@example.com"`                  // 账号邮箱
      CreatedAt time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`        // 创建时间
  }
  ```

#### Task 3.2: Order Request 模型注释

**文件：`dpmain/internal/app/domains/apimodel/request/create_order_request.go`**

按顺序为以下结构体添加字段注释和示例：

- [ ] `CreateOrderRequest`
  ```go
  // CreateOrderRequest 创建订单请求
  type CreateOrderRequest struct {
      AccountID       int64     `json:"account_id" binding:"required" example:"1"`                    // 账号ID
      MerchantOrderNo string    `json:"merchant_order_no" binding:"required" example:"ORD-20240101-001"` // 商户订单号（唯一）
      Shipment        *Shipment `json:"shipment" binding:"required"`                                  // 货件信息
  }
  ```

- [ ] `Shipment`
  ```go
  // Shipment 货件信息
  type Shipment struct {
      ShipFrom *Address  `json:"ship_from" binding:"required"` // 发货地址
      ShipTo   *Address  `json:"ship_to" binding:"required"`   // 收货地址
      Parcels  []*Parcel `json:"parcels" binding:"required"`   // 包裹列表（至少1个）
  }
  ```

- [ ] `Address`
  ```go
  // Address 地址信息
  type Address struct {
      ContactName string `json:"contact_name" binding:"required" example:"John Doe"`         // 联系人姓名
      CompanyName string `json:"company_name" example:"ACME Corp"`                           // 公司名称（可选）
      Street1     string `json:"street1" binding:"required" example:"123 Main St"`          // 地址行1
      Street2     string `json:"street2" example:"Suite 100"`                                // 地址行2（可选）
      City        string `json:"city" binding:"required" example:"San Francisco"`           // 城市
      State       string `json:"state" example:"CA"`                                         // 州/省（可选）
      PostalCode  string `json:"postal_code" binding:"required" example:"94102"`            // 邮政编码
      Country     string `json:"country" binding:"required" example:"USA"`                  // 国家（ISO 3166-1 alpha-3）
      Phone       string `json:"phone" example:"+1-415-555-0100"`                            // 联系电话（可选）
      Email       string `json:"email" example:"john@example.com"`                           // 联系邮箱（可选）
  }
  ```

- [ ] `Parcel`
  ```go
  // Parcel 包裹信息
  type Parcel struct {
      Weight    *Weight    `json:"weight" binding:"required"`    // 重量
      Dimension *Dimension `json:"dimension"`                    // 尺寸（可选）
      Items     []*Item    `json:"items" binding:"required"`     // 商品列表（至少1个）
  }
  ```

- [ ] `Weight`
  ```go
  // Weight 重量信息
  type Weight struct {
      Value float64 `json:"value" binding:"required" example:"1.5"`  // 重量值（必须 > 0）
      Unit  string  `json:"unit" binding:"required" example:"kg"`    // 重量单位（kg, lb）
  }
  ```

- [ ] `Dimension`
  ```go
  // Dimension 尺寸信息
  type Dimension struct {
      Width  float64 `json:"width" example:"10.0"`   // 宽度
      Height float64 `json:"height" example:"20.0"`  // 高度
      Depth  float64 `json:"depth" example:"15.0"`   // 深度
      Unit   string  `json:"unit" example:"cm"`      // 尺寸单位（cm, in）
  }
  ```

- [ ] `Item`
  ```go
  // Item 商品信息
  type Item struct {
      Description string  `json:"description" binding:"required" example:"T-Shirt"` // 商品描述
      Quantity    int     `json:"quantity" binding:"required" example:"2"`          // 数量（必须 > 0）
      Price       *Money  `json:"price" binding:"required"`                         // 单价
      SKU         string  `json:"sku" example:"TSH-001"`                            // SKU 编码（可选）
      Weight      *Weight `json:"weight"`                                           // 单件重量（可选）
  }
  ```

- [ ] `Money`
  ```go
  // Money 金额信息
  type Money struct {
      Amount   float64 `json:"amount" binding:"required" example:"19.99"` // 金额
      Currency string  `json:"currency" binding:"required" example:"USD"` // 货币代码（ISO 4217）
  }
  ```

#### Task 3.3: Order Response 模型注释

**文件：`dpmain/internal/app/domains/apimodel/response/order_response.go`**

- [ ] `OrderResponse`
  ```go
  // OrderResponse 订单响应
  type OrderResponse struct {
      ID              string           `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`      // 订单ID（UUID）
      AccountID       int64            `json:"account_id" example:"1"`                                  // 账号ID
      MerchantOrderNo string           `json:"merchant_order_no" example:"ORD-20240101-001"`           // 商户订单号
      Status          string           `json:"status" example:"COMPLETED" enums:"PENDING,DIAGNOSING,COMPLETED,FAILED"` // 订单状态
      Diagnosis       *DiagnosisResult `json:"diagnosis,omitempty"`                                     // 诊断结果（可选）
      CreatedAt       time.Time        `json:"created_at" example:"2024-01-01T00:00:00Z"`              // 创建时间
      UpdatedAt       time.Time        `json:"updated_at" example:"2024-01-01T00:00:00Z"`              // 更新时间
  }
  ```

- [ ] `DiagnosisResult`
  ```go
  // DiagnosisResult 诊断结果
  type DiagnosisResult struct {
      Items []*DiagnosisItem `json:"items"` // 诊断项列表
  }
  ```

- [ ] `DiagnosisItem`
  ```go
  // DiagnosisItem 诊断项
  type DiagnosisItem struct {
      Type     string      `json:"type" example:"shipping" enums:"shipping,anomaly"`     // 诊断类型
      Status   string      `json:"status" example:"SUCCESS" enums:"SUCCESS,FAILED"`      // 诊断状态
      DataJSON interface{} `json:"data_json"`                                            // 诊断数据（JSON）
      Error    string      `json:"error,omitempty" example:""`                           // 错误信息（可选）
  }
  ```

#### Task 3.4: 统一 Response 模型注释

**文件：`dpmain/internal/app/pkg/ginx/response.go`**

- [ ] `Response`
  ```go
  // Response 统一响应结构
  type Response struct {
      Code    int         `json:"code" example:"200"`                      // 业务状态码（200=成功, 3001=处理中, 400=客户端错误, 500=服务器错误）
      Message string      `json:"message,omitempty" example:"success"`     // 响应消息（错误时返回）
      Data    interface{} `json:"data,omitempty"`                          // 响应数据
      PollURL string      `json:"poll_url,omitempty" example:"/api/v1/orders/550e8400-e29b-41d4-a716-446655440000"` // 轮询URL（仅 code=3001 时返回）
  }
  ```

---

### 阶段 4: 添加 Handler 接口注释（预计 60 分钟）

#### Task 4.1: Health Check 接口

**文件：`dpmain/internal/app/server/routers/router.go`**

- [ ] 修改 `/health` 路由注册为独立函数：
  ```go
  // HealthCheck 健康检查
  // @Summary      健康检查
  // @Description  检查服务运行状态
  // @Tags         system
  // @Produce      json
  // @Success      200 {object} map[string]string
  // @Router       /health [get]
  func HealthCheck(c *gin.Context) {
      c.JSON(200, gin.H{
          "status":  "ok",
          "service": "dpmain",
          "message": "Service is running",
      })
  }
  ```
- [ ] 修改路由注册：`r.GET("/health", HealthCheck)`

#### Task 4.2: Account 接口注释

**文件：`dpmain/internal/app/server/handlers/account/create.go`**

- [ ] 修改 `Create` 方法注释：
  ```go
  // Create 创建账号
  // @Summary      创建账号
  // @Description  创建一个新的账号，用于后续订单关联
  // @Tags         accounts
  // @Accept       json
  // @Produce      json
  // @Param        request body request.CreateAccountRequest true "创建账号请求"
  // @Success      200 {object} ginx.Response{data=response.AccountResponse} "创建成功"
  // @Failure      400 {object} ginx.Response "参数错误"
  // @Failure      500 {object} ginx.Response "服务器错误"
  // @Security     ApiKeyAuth
  // @Router       /accounts [post]
  func (h *AccountHandler) Create(c *gin.Context) {
      // ... 现有代码
  }
  ```

**文件：`dpmain/internal/app/server/handlers/account/get.go`**

- [ ] 修改 `Get` 方法注释：
  ```go
  // Get 获取账号详情
  // @Summary      获取账号详情
  // @Description  根据账号ID获取账号详细信息
  // @Tags         accounts
  // @Produce      json
  // @Param        id path int true "账号ID"
  // @Success      200 {object} ginx.Response{data=response.AccountResponse} "查询成功"
  // @Failure      400 {object} ginx.Response "参数错误"
  // @Failure      404 {object} ginx.Response "账号不存在"
  // @Failure      500 {object} ginx.Response "服务器错误"
  // @Security     ApiKeyAuth
  // @Router       /accounts/{id} [get]
  func (h *AccountHandler) Get(c *gin.Context) {
      // ... 现有代码
  }
  ```

#### Task 4.3: Order 接口注释

**文件：`dpmain/internal/app/server/handlers/order/create.go`**

- [ ] 修改 `Create` 方法注释（重点接口，需详细说明）：
  ```go
  // Create 创建订单
  // @Summary      创建订单
  // @Description  创建订单并触发智能诊断（物流费率计算 + 异常检测）
  // @Description
  // @Description  **Smart Wait 机制说明：**
  // @Description  - 接口会 Hold 10s 等待诊断结果
  // @Description  - 10s 内完成诊断：返回 200 OK，包含完整诊断结果
  // @Description  - 10s 超时：返回 200 OK，code=3001（Processing），需要通过 poll_url 轮询结果
  // @Description
  // @Description  **订单状态说明：**
  // @Description  - PENDING: 订单已创建，等待诊断
  // @Description  - DIAGNOSING: 诊断进行中
  // @Description  - COMPLETED: 诊断完成（成功或失败）
  // @Description  - FAILED: 订单处理失败
  // @Tags         orders
  // @Accept       json
  // @Produce      json
  // @Param        request body request.CreateOrderRequest true "创建订单请求"
  // @Success      200 {object} ginx.Response{data=response.OrderResponse} "创建成功（诊断完成）"
  // @Success      200 {object} ginx.Response{code=3001,poll_url=string} "创建成功（诊断进行中，需轮询）"
  // @Failure      400 {object} ginx.Response "参数错误"
  // @Failure      500 {object} ginx.Response "服务器错误"
  // @Security     ApiKeyAuth
  // @Router       /orders [post]
  func (h *OrderHandler) Create(c *gin.Context) {
      // ... 现有代码
  }
  ```

**文件：`dpmain/internal/app/server/handlers/order/get.go`**

- [ ] 修改 `Get` 方法注释：
  ```go
  // Get 获取订单详情
  // @Summary      获取订单详情
  // @Description  根据订单ID获取订单详细信息（包含诊断结果）
  // @Description
  // @Description  **使用场景：**
  // @Description  - 创建订单返回 code=3001 时，通过此接口轮询结果
  // @Description  - 查询历史订单详情
  // @Tags         orders
  // @Produce      json
  // @Param        id path string true "订单ID（UUID）"
  // @Success      200 {object} ginx.Response{data=response.OrderResponse} "查询成功"
  // @Failure      400 {object} ginx.Response "参数错误"
  // @Failure      404 {object} ginx.Response "订单不存在"
  // @Failure      500 {object} ginx.Response "服务器错误"
  // @Security     ApiKeyAuth
  // @Router       /orders/{id} [get]
  func (h *OrderHandler) Get(c *gin.Context) {
      // ... 现有代码
  }
  ```

---

### 阶段 5: 生成和测试 Swagger 文档（预计 30 分钟）

#### Task 5.1: 生成 Swagger 文档
- [ ] 运行：`make swagger`
- [ ] 检查生成文件：
  - `dpmain/docs/docs.go`
  - `dpmain/docs/swagger.json`
  - `dpmain/docs/swagger.yaml`
- [ ] 检查生成日志，确认无错误和警告

#### Task 5.2: 启动服务并访问文档
- [ ] 确保依赖服务已启动（MySQL, Redis, Lmstfy）
- [ ] 启动 dpmain 服务：`make run-dpmain`
- [ ] 浏览器访问：`http://localhost:8080/swagger/index.html`
- [ ] 验证 Swagger UI 正确加载

#### Task 5.3: 功能验证

**基础验证：**
- [ ] 文档标题显示为 "OIP Backend API v1.0"
- [ ] 所有 5 个接口正确显示
- [ ] 接口按 Tags 正确分组：
  - system (1个)
  - accounts (2个)
  - orders (2个)

**详细验证：**
- [ ] 展开每个接口，检查：
  - [ ] Summary 和 Description 完整
  - [ ] Parameters 正确显示
  - [ ] Request Body Schema 完整（POST 接口）
  - [ ] Responses 各状态码都有说明
  - [ ] Example Value 可正常显示

**交互验证：**
- [ ] 点击 "Try it out" 按钮
- [ ] 测试创建账号接口：
  - 输入示例数据
  - Execute 发送请求
  - 检查响应结果
- [ ] 测试创建订单接口：
  - 使用生成的 account_id
  - 输入完整 shipment 数据
  - 检查返回的订单状态

**Models 验证：**
- [ ] 滚动到页面底部 "Schemas" 部分
- [ ] 检查所有模型定义是否完整：
  - CreateAccountRequest
  - CreateOrderRequest
  - Shipment, Address, Parcel, Weight, Dimension, Item, Money
  - Response
  - AccountResponse
  - OrderResponse, DiagnosisResult, DiagnosisItem

#### Task 5.4: 文档导出测试
- [ ] 在 Swagger UI 顶部找到 "Download" 链接
- [ ] 下载 `swagger.json` 并检查格式
- [ ] 下载 `swagger.yaml` 并检查格式
- [ ] 尝试导入 Postman：
  - 打开 Postman
  - Import → Upload Files
  - 选择 `swagger.json`
  - 验证所有接口正确导入

---

### 阶段 6: 文档和代码质量（预计 30 分钟）

#### Task 6.1: 更新 dpmain README

**文件：`dpmain/README.md`**

- [ ] 添加 "API 文档" 章节：
  ```markdown
  ## API 文档
  
  ### Swagger 文档
  
  本项目使用 Swagger（OpenAPI 3.0）生成 API 文档。
  
  #### 生成文档
  
  ```bash
  # 在项目根目录执行
  make swagger
  ```

  #### 访问文档

  启动服务后访问：
  - **Swagger UI**: http://localhost:8080/swagger/index.html
  - **JSON 格式**: http://localhost:8080/swagger/doc.json

  #### 快速开始

  1. 启动服务：
     ```bash
     make run-dpmain
     ```

  2. 打开浏览器访问 Swagger UI

  3. 点击 "Authorize" 按钮，输入 API Key（当前版本暂未启用认证，可跳过）

  4. 选择任意接口，点击 "Try it out" 进行测试

  #### API 概览

  - **Base URL**: `http://localhost:8080/api/v1`
  - **认证方式**: API Key (Header: `api-key`) - 暂未启用
  - **文档版本**: v1.0

  #### 接口列表

  | 分组 | 方法 | 路径 | 说明 |
  |------|------|------|------|
  | System | GET | `/health` | 健康检查 |
  | Accounts | POST | `/api/v1/accounts` | 创建账号 |
  | Accounts | GET | `/api/v1/accounts/{id}` | 获取账号详情 |
  | Orders | POST | `/api/v1/orders` | 创建订单（触发诊断） |
  | Orders | GET | `/api/v1/orders/{id}` | 获取订单详情 |

  #### 快速测试示例

  **创建账号：**
  ```bash
  curl -X POST http://localhost:8080/api/v1/accounts \
    -H "Content-Type: application/json" \
    -H "api-key: your-api-key" \
    -d '{
      "name": "John Doe",
      "email": "john@example.com"
    }'
  ```

  **创建订单：**
  ```bash
  curl -X POST http://localhost:8080/api/v1/orders \
    -H "Content-Type: application/json" \
    -H "api-key: your-api-key" \
    -d '{
      "account_id": 1,
      "merchant_order_no": "ORD-20240101-001",
      "shipment": {
        "ship_from": {
          "contact_name": "Seller Store",
          "company_name": "ACME Corp",
          "street1": "123 Main St",
          "city": "San Francisco",
          "state": "CA",
          "postal_code": "94102",
          "country": "USA",
          "phone": "+1-415-555-0100",
          "email": "seller@example.com"
        },
        "ship_to": {
          "contact_name": "John Doe",
          "street1": "456 Oak Ave",
          "city": "Los Angeles",
          "state": "CA",
          "postal_code": "90001",
          "country": "USA",
          "phone": "+1-213-555-0200",
          "email": "buyer@example.com"
        },
        "parcels": [
          {
            "weight": {
              "value": 1.5,
              "unit": "kg"
            },
            "dimension": {
              "width": 10.0,
              "height": 20.0,
              "depth": 15.0,
              "unit": "cm"
            },
            "items": [
              {
                "description": "T-Shirt",
                "quantity": 2,
                "price": {
                  "amount": 19.99,
                  "currency": "USD"
                },
                "sku": "TSH-001"
              }
            ]
          }
        ]
      }
    }'
  ```

  #### 导出文档

  Swagger 文档可导出为多种格式：
  - 通过 Swagger UI 下载 JSON/YAML
  - 导入 Postman / Insomnia 等 API 工具
  - 集成到 CI/CD 流程

  #### 注意事项

  1. **Smart Wait 机制**：创建订单接口会等待 10s 等待诊断结果
     - 10s 内完成：返回完整诊断结果
     - 10s 超时：返回 code=3001，需通过 GET 接口轮询

  2. **订单状态**：
     - `PENDING`: 等待诊断
     - `DIAGNOSING`: 诊断进行中
     - `COMPLETED`: 诊断完成
     - `FAILED`: 处理失败

  3. **认证**：当前版本 API Key 认证暂未启用，后续版本会完善
  ```

#### Task 6.2: 代码质量检查
- [ ] 运行格式化：
  ```bash
  cd /Users/cooperswang/Documents/wsqyouth/oip/backend/dpmain
  gofmt -w .
  ```
- [ ] 运行静态检查：
  ```bash
  go vet ./...
  ```
- [ ] 重新编译验证：
  ```bash
  go build -o bin/dpmain cmd/apiserver/main.go
  ```
- [ ] 确认无编译错误

#### Task 6.3: 清理和确认
- [ ] 检查 git 状态，确认修改文件列表
- [ ] 删除临时文件（如有）
- [ ] 确认 `.gitignore` 不忽略 `docs/` 目录
- [ ] 最终测试：
  ```bash
  make swagger
  make run-dpmain
  # 访问 http://localhost:8080/swagger/index.html
  ```

---

### 阶段 7: 交付和验收（预计 20 分钟）

#### Task 7.1: 完整回归测试
- [ ] 重启所有服务（MySQL, Redis, Lmstfy, dpmain）
- [ ] 访问 Swagger UI，逐个测试所有接口
- [ ] 验证文档和实际接口行为一致

#### Task 7.2: 准备演示
- [ ] 准备截图：
  - Swagger UI 首页
  - Accounts 接口详情
  - Orders 接口详情（展示 Try it out）
  - Models Schema 列表
- [ ] 录制快速演示视频（可选）

#### Task 7.3: 文档归档
- [ ] 确认 `docs/` 目录文件完整
- [ ] 检查 README 更新是否完整
- [ ] 更新 Story 文件状态（story-002-生成swagger文档.md）

#### Task 7.4: 代码提交
- [ ] 使用项目自动化提交脚本：
  ```bash
  cd /Users/cooperswang/Documents/wsqyouth/oip/backend
  ../.claude/commands/commit.sh "feat: add swagger documentation for all APIs" --rebase
  ```
- [ ] 推送到远程仓库（如需要）

---

## 🔧 技术方案

### 工具选型
- **Swagger 生成器**: `swaggo/swag` - Go 官方推荐的 Swagger 工具
- **Gin 集成**: `gin-swagger` - Gin 框架 Swagger 中间件
- **规范版本**: OpenAPI 3.0

### 注释规范
基于 `swaggo/swag` 的注释语法：
- `@title`: API 标题
- `@version`: API 版本
- `@description`: API 描述
- `@host`: 主机地址
- `@BasePath`: 基础路径
- `@Summary`: 接口简述
- `@Description`: 接口详细说明
- `@Tags`: 接口分组
- `@Accept`: 接受的 Content-Type
- `@Produce`: 返回的 Content-Type
- `@Param`: 参数定义
- `@Success`: 成功响应
- `@Failure`: 失败响应
- `@Security`: 认证方式
- `@Router`: 路由定义

### 目录结构
```
dpmain/
├── docs/                           # Swagger 生成文件（新增）
│   ├── docs.go                     # Go 代码
│   ├── swagger.json                # JSON 格式
│   └── swagger.yaml                # YAML 格式
├── cmd/apiserver/
│   └── main.go                     # 添加 API 总体注释
├── internal/app/
│   ├── domains/apimodel/
│   │   ├── request/               # 完善字段注释和示例
│   │   └── response/              # 完善字段注释和示例
│   └── server/
│       ├── routers/router.go      # 添加 Swagger 路由
│       └── handlers/              # 添加接口注释
└── README.md                       # 更新使用说明
```

---

## 📊 验收标准

### 必须满足（Must Have）
- [x] 所有 5 个接口都有完整的 Swagger 文档
- [x] 每个字段都有类型、required、description 标注
- [x] 可通过 `http://localhost:8080/swagger/index.html` 访问
- [x] Swagger UI 可正常渲染和交互
- [x] 代码通过 `gofmt` 和 `go vet` 检查
- [x] README 包含 Swagger 使用说明

### 应该满足（Should Have）
- [x] 提供完整的请求/响应示例
- [x] Models 独立定义并可查看
- [x] Try it out 功能可用
- [x] 可导出 JSON/YAML 格式

### 可以满足（Could Have）
- [ ] 提供 Postman Collection 导出
- [ ] 添加接口性能说明
- [ ] 提供常见错误码文档

---

## 📚 参考资料

### 官方文档
- **swaggo/swag**: https://github.com/swaggo/swag
- **gin-swagger**: https://github.com/swaggo/gin-swagger
- **OpenAPI Specification**: https://swagger.io/specification/

### 示例项目
- AfterShip API Docs: https://www.aftership.com/docs/shipping/
- Gin Swagger Example: https://github.com/swaggo/swag/tree/master/example/celler

### 注释规范
- **声明式注释**: https://github.com/swaggo/swag#declarative-comments-format
- **参数类型**: https://github.com/swaggo/swag#param-type
- **数据类型**: https://github.com/swaggo/swag#data-type

---

## ⏱️ 时间估算

| 阶段 | 预计时间 |
|------|----------|
| 阶段 1: 环境准备 | 30 分钟 |
| 阶段 2: API 总体配置 | 20 分钟 |
| 阶段 3: Request/Response 注释 | 40 分钟 |
| 阶段 4: Handler 接口注释 | 60 分钟 |
| 阶段 5: 生成和测试 | 30 分钟 |
| 阶段 6: 文档和代码质量 | 30 分钟 |
| 阶段 7: 交付和验收 | 20 分钟 |
| **总计** | **约 3.5 小时** |

---

## 🎯 下一步行动

1. ✅ 阅读本 PRD，确认理解所有任务
2. ⏸️ 等待开发者确认开始开发
3. 🚀 按阶段顺序执行任务
4. ✅ 每完成一个阶段，标记对应的 checklist
5. 📝 遇到问题记录到"遇到的问题与解决方案"
6. 🎉 完成后更新 Story 状态并提交代码

---

## 📝 备注

- **API Key 名称**: 使用 `api-key` 而非 `as-api-key`
- **当前版本**: 认证暂未实现，仅作为文档占位
- **Smart Wait**: 订单创建接口的特殊机制，需在文档中重点说明
- **文档风格**: 参考 AfterShip，清晰、专业、易用

---

**最后更新**: 2026-01-07
**PRD 版本**: v1.0
