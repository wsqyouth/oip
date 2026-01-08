package request

// CreateAccountRequest 创建账号请求
type CreateAccountRequest struct {
	Name  string `json:"name" binding:"required" example:"John Doe"`
	Email string `json:"email" binding:"required,email" example:"john@example.com"`
}
