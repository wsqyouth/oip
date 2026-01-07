package ginx

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	PollURL string      `json:"poll_url,omitempty"`
}

// Success 成功响应（200）
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: 200,
		Data: data,
	})
}

// Error 错误响应（400/500）
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}

// Processing 处理中响应（3001），用于 Smart Wait 超时场景
// 明确判断：订单状态为 DIAGNOSING 时返回此响应
func Processing(c *gin.Context, orderID string, pollURL string) {
	c.JSON(http.StatusOK, Response{
		Code:    3001,
		Message: "Order is being diagnosed, please poll for results",
		Data: map[string]string{
			"order_id": orderID,
		},
		PollURL: pollURL,
	})
}

// BadRequest 400 错误
func BadRequest(c *gin.Context, message string) {
	Error(c, http.StatusBadRequest, message)
}

// NotFound 404 错误
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, message)
}

// InternalError 500 错误
func InternalError(c *gin.Context, message string) {
	Error(c, http.StatusInternalServerError, message)
}
