package request

// CreateAccountRequest 创建账号请求（DTO）
type CreateAccountRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}
