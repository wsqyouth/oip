package errorutil

import "fmt"

// Error 错误结构（包含可重试标记）
type Error struct {
	Code       int    `json:"code"`
	Message    string `json:"message"`
	Retryable  bool   `json:"retryable"`
	DevDetails string `json:"dev_details,omitempty"`
}

// Error 实现 error 接口
func (e *Error) Error() string {
	return e.Message
}

// Retriable 创建可重试错误（网络错误、临时故障等）
func Retriable(message string) *Error {
	return &Error{
		Code:      500,
		Message:   message,
		Retryable: true,
	}
}

// RetriableWithDetails 创建可重试错误（带详细信息）
func RetriableWithDetails(message string, details string) *Error {
	return &Error{
		Code:       500,
		Message:    message,
		Retryable:  true,
		DevDetails: details,
	}
}

// NonRetriable 创建不可重试错误（参数错误、业务规则错误等）
func NonRetriable(message string) *Error {
	return &Error{
		Code:      400,
		Message:   message,
		Retryable: false,
	}
}

// NonRetriableWithDetails 创建不可重试错误（带详细信息）
func NonRetriableWithDetails(message string, details string) *Error {
	return &Error{
		Code:       400,
		Message:    message,
		Retryable:  false,
		DevDetails: details,
	}
}

// Wrap 包装错误（自动判断是否可重试）
func Wrap(err error) *Error {
	if err == nil {
		return nil
	}

	// 如果已经是 Error 类型，直接返回
	if e, ok := err.(*Error); ok {
		return e
	}

	// 默认为不可重试错误
	return &Error{
		Code:       500,
		Message:    err.Error(),
		Retryable:  false,
		DevDetails: fmt.Sprintf("%+v", err),
	}
}

// UnWrapResponse 解包错误（用于 Response）
func UnWrapResponse(err error) *Error {
	if err == nil {
		return nil
	}
	return Wrap(err)
}
