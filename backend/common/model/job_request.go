package model

// OrderDiagnoseJob 订单诊断任务消息（标准化）
// 用于 dpmain → dpsync 的消息传递
type OrderDiagnoseJob struct {
	Payload OrderDiagnosePayload `json:"payload"`
}

// OrderDiagnosePayload Job 负载
type OrderDiagnosePayload struct {
	Data OrderDiagnoseData `json:"data"`
}

// OrderDiagnoseData Job 数据层
type OrderDiagnoseData struct {
	// 元信息
	RequestID  string `json:"request_id"`  // 请求 ID（全链路追踪）
	OrgID      string `json:"org_id"`      // 组织 ID（MVP 固定为 "0"）
	ActionType string `json:"action_type"` // 动作类型，固定值 "order_diagnose"
	ID         string `json:"id"`          // 订单 ID

	// 业务数据
	Data OrderDiagnoseBusinessData `json:"data"`
}

// OrderDiagnoseBusinessData 订单诊断业务数据
// 包含 dpsync 执行诊断所需的所有数据（避免查询 DB）
type OrderDiagnoseBusinessData struct {
	OrderID         string                 `json:"order_id"`          // 订单 ID
	AccountID       int64                  `json:"account_id"`        // 账户 ID
	MerchantOrderNo string                 `json:"merchant_order_no"` // 商家订单号
	Shipment        map[string]interface{} `json:"shipment"`          // 物流信息（TBC: 未来可定义为具体结构体）
}
