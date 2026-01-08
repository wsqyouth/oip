package ginx

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// Response 统一响应结构
type Response struct {
	Meta Meta        `json:"meta"`
	Data interface{} `json:"data,omitempty"`
}

// Meta 元数据
type Meta struct {
	Code    int           `json:"code" example:"200"`
	Message string        `json:"message" example:"OK"`
	Details []ErrorDetail `json:"details,omitempty"`
}

// ErrorDetail 错误详情
type ErrorDetail struct {
	Path string `json:"path" example:"email"`
	Info string `json:"info" example:"email is required"`
}

// ProcessingData Smart Wait 超时返回的数据
type ProcessingData struct {
	OrderID string `json:"order_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	PollURL string `json:"poll_url" example:"/api/v1/orders/550e8400-e29b-41d4-a716-446655440000"`
}

// Success 成功响应（200）
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Meta: Meta{
			Code:    200,
			Message: "OK",
		},
		Data: data,
	})
}

// Error 错误响应（400/500）
func Error(c *gin.Context, httpCode int, message string) {
	c.JSON(httpCode, Response{
		Meta: Meta{
			Code:    httpCode,
			Message: message,
		},
	})
}

// ErrorWithDetails 带详情的错误响应
func ErrorWithDetails(c *gin.Context, httpCode int, message string, details []ErrorDetail) {
	c.JSON(httpCode, Response{
		Meta: Meta{
			Code:    httpCode,
			Message: message,
			Details: details,
		},
	})
}

// Processing 处理中响应（3001），用于 Smart Wait 超时场景
func Processing(c *gin.Context, orderID string, pollURL string) {
	c.JSON(http.StatusOK, Response{
		Meta: Meta{
			Code:    3001,
			Message: "Order is being diagnosed, please poll for results",
		},
		Data: ProcessingData{
			OrderID: orderID,
			PollURL: pollURL,
		},
	})
}

// BadRequest 400 错误
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// BadRequestWithValidation 400 错误（带验证详情）
func BadRequestWithValidation(c *gin.Context, err error) {
	if validationErrs, ok := err.(validator.ValidationErrors); ok {
		details := make([]ErrorDetail, 0, len(validationErrs))
		for _, fieldErr := range validationErrs {
			details = append(details, ErrorDetail{
				Path: fieldErr.Field(),
				Info: getValidationErrorMessage(fieldErr),
			})
		}
		ErrorWithDetails(c, http.StatusBadRequest, "Validation failed", details)
		return
	}

	BadRequest(c, err.Error())
}

// NotFound 404 错误
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// InternalError 500 错误
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}

// getValidationErrorMessage 根据验证错误类型返回友好的错误消息
func getValidationErrorMessage(fieldErr validator.FieldError) string {
	switch fieldErr.Tag() {
	case "required":
		return fieldErr.Field() + " is required"
	case "email":
		return fieldErr.Field() + " must be a valid email address"
	case "min":
		return fieldErr.Field() + " must be at least " + fieldErr.Param()
	case "max":
		return fieldErr.Field() + " must be at most " + fieldErr.Param()
	default:
		return fieldErr.Field() + " is invalid"
	}
}
