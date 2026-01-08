package account

import (
	"log"

	"github.com/gin-gonic/gin"
	"oip/dpmain/internal/app/domains/apimodel/request"
	"oip/dpmain/internal/app/domains/apimodel/response"
	"oip/dpmain/internal/app/pkg/ginx"
)

// Create godoc
// @Summary      创建账号
// @Description  创建一个新的账号，用于后续订单关联
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        request body request.CreateAccountRequest true "创建账号请求"
// @Success      200 {object} ginx.Response{data=response.AccountResponse} "创建成功"
// @Failure      400 {object} ginx.Response "参数错误"
// @Failure      500 {object} ginx.Response "服务器错误"
// @Security     ApiKeyAuth
// @Router       /accounts [post]
func (h *AccountHandler) Create(c *gin.Context) {
	var req request.CreateAccountRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ginx.BadRequestWithValidation(c, err)
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
