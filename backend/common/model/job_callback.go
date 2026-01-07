package model

// OrderDiagnoseCallback 订单诊断回调消息（标准化）
// 用于 dpsync → dpmain callback consumer 的消息传递
type OrderDiagnoseCallback struct {
	RequestID       string               `json:"request_id"`                 // 对应请求的 request_id（链路追踪）
	OrderID         string               `json:"order_id"`                   // 订单 ID
	AccountID       int64                `json:"account_id"`                 // 账户 ID
	Status          string               `json:"status"`                     // 回调状态: SUCCESS / FAILED
	DiagnosisResult *DiagnosisResultData `json:"diagnosis_result,omitempty"` // 诊断结果（成功时返回）
	Error           string               `json:"error,omitempty"`            // 错误信息（失败时返回）
	ProcessedAt     int64                `json:"processed_at"`               // 处理时间戳（Unix timestamp）
}

// 回调状态常量
const (
	CallbackStatusSuccess = "SUCCESS" // 诊断成功
	CallbackStatusFailed  = "FAILED"  // 诊断失败
)
