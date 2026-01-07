package model

// Response 统一响应结构
type Response struct {
	Meta MetaInfo    `json:"meta"`
	Data interface{} `json:"data,omitempty"`
}

// MetaInfo 响应元信息
type MetaInfo struct {
	Code    int           `json:"code"`
	Type    string        `json:"type"`
	Message string        `json:"message"`
	Details []ErrorDetail `json:"details,omitempty"`
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	Path string `json:"path"`
	Info string `json:"info"`
}

// 响应类型常量
const (
	ResponseTypeOK              = "OK"
	ResponseTypeValidationError = "ValidationError"
	ResponseTypeNotFound        = "NotFound"
	ResponseTypeInternalError   = "InternalError"
	ResponseTypeProcessing      = "Processing"
)
