package errorx

import "errors"

// 定义业务错误（预留）
var (
	ErrAccountNotFound  = errors.New("account not found")
	ErrOrderNotFound    = errors.New("order not found")
	ErrDuplicateOrder   = errors.New("duplicate order")
	ErrInvalidAddress   = errors.New("invalid address format")
	ErrDiagnosisTimeout = errors.New("diagnosis timeout")
)

// BusinessError 业务错误结构
type BusinessError struct {
	Code    int
	Message string
	Details []ErrorDetail
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	Path string
	Info string
}

// Error 实现 error 接口
func (e *BusinessError) Error() string {
	return e.Message
}

// NewBusinessError 创建业务错误
func NewBusinessError(code int, message string) *BusinessError {
	return &BusinessError{
		Code:    code,
		Message: message,
	}
}
