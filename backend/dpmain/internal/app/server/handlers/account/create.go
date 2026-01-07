package account

import (
	"log"

	"github.com/gin-gonic/gin"
	"oip/dpmain/internal/app/domains/apimodel/request"
	"oip/dpmain/internal/app/domains/apimodel/response"
	"oip/dpmain/internal/app/pkg/ginx"
)

// Create 创建账号接口
// POST /api/v1/accounts
func (h *AccountHandler) Create(c *gin.Context) {
	var req request.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.BadRequest(c, err.Error())
		return
	}

	account, err := h.accountService.CreateAccount(c.Request.Context(), req.Name, req.Email)
	if err != nil {
		log.Printf("[ERROR] create account failed: %v", err)
		ginx.InternalError(c, err.Error())
		return
	}

	ginx.Success(c, response.FromAccountEntity(account))
}
