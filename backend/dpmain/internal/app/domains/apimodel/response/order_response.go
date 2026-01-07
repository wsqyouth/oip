package response

import "time"

// OrderResponse 订单响应（DTO）
type OrderResponse struct {
	ID              string           `json:"id"`
	AccountID       int64            `json:"account_id"`
	MerchantOrderNo string           `json:"merchant_order_no"`
	Status          string           `json:"status"`
	Diagnosis       *DiagnosisResult `json:"diagnosis,omitempty"`
	CreatedAt       time.Time        `json:"created_at"`
	UpdatedAt       time.Time        `json:"updated_at"`
}

// DiagnosisResult 诊断结果（DTO）
type DiagnosisResult struct {
	Items []*DiagnosisItem `json:"items"`
}

// DiagnosisItem 诊断项（DTO）
type DiagnosisItem struct {
	Type     string      `json:"type"`
	Status   string      `json:"status"`
	DataJSON interface{} `json:"data_json"`
	Error    string      `json:"error,omitempty"`
}
