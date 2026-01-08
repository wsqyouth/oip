package response

import "time"

// OrderResponse 订单响应
type OrderResponse struct {
	ID              string           `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	AccountID       int64            `json:"account_id" example:"1"`
	MerchantOrderNo string           `json:"merchant_order_no" example:"ORD-20240101-001"`
	Status          string           `json:"status" example:"COMPLETED" enums:"PENDING,DIAGNOSING,COMPLETED,FAILED"`
	Diagnosis       *DiagnosisResult `json:"diagnosis,omitempty"`
	CreatedAt       time.Time        `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt       time.Time        `json:"updated_at" example:"2024-01-01T00:00:00Z"`
}

// DiagnosisResult 诊断结果
type DiagnosisResult struct {
	Items []*DiagnosisItem `json:"items"`
}

// DiagnosisItem 诊断项
type DiagnosisItem struct {
	Type     string      `json:"type" example:"shipping" enums:"shipping,anomaly"`
	Status   string      `json:"status" example:"SUCCESS" enums:"SUCCESS,FAILED"`
	DataJSON interface{} `json:"data_json"`
	Error    string      `json:"error,omitempty" example:""`
}
