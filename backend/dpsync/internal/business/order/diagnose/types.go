package diagnose

import "oip/common/model"

// DiagnosePayload Job 消息中的业务数据
type DiagnosePayload struct {
	OrderID         string                 `json:"order_id"`
	AccountID       int64                  `json:"account_id"`
	MerchantOrderNo string                 `json:"merchant_order_no"`
	Shipment        map[string]interface{} `json:"shipment"`
}

// DiagnoseInput 诊断服务输入
type DiagnoseInput struct {
	RequestID       string
	OrderID         string
	AccountID       int64
	MerchantOrderNo string
	Shipment        map[string]interface{}
}

// DiagnosisResultData 业务处理结果
type DiagnosisResultData struct {
	Items       []model.DiagnosisItem
	OrderID     string
	ProcessedAt int64
}

// DiagnosisOutput 最终输出结构
type DiagnosisOutput struct {
	Items       []model.DiagnosisItem `json:"items"`
	OrderID     string                `json:"order_id"`
	ProcessedAt int64                 `json:"processed_at"`
}
