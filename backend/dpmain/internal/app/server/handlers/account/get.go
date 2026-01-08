package account

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"oip/dpmain/internal/app/domains/apimodel/response"
	"oip/dpmain/internal/app/pkg/ginx"
)

// Get godoc
// @Summary      获取账号详情
// @Description  根据账号ID获取账号详细信息
// @Tags         accounts
// @Produce      json
// @Param        id path int true "账号ID"
// @Success      200 {object} ginx.Response{data=response.AccountResponse} "查询成功"
// @Failure      400 {object} ginx.Response "参数错误"
// @Failure      404 {object} ginx.Response "账号不存在"
// @Failure      500 {object} ginx.Response "服务器错误"
// @Security     ApiKeyAuth
// @Router       /accounts/{id} [get]
func (h *AccountHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	accountID, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		ginx.BadRequest(c, "invalid account_id")
		return
	}

	account, err := h.accountService.GetAccount(c.Request.Context(), accountID)
	if err != nil {
		log.Printf("[ERROR] get account failed: %v", err)
		ginx.NotFound(c, "account not found")
		return
	}

	ginx.Success(c, response.FromAccountEntity(account))
}
