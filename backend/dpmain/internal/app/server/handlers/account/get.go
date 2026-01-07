package account

import (
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
	"oip/dpmain/internal/app/domains/apimodel/response"
	"oip/dpmain/internal/app/pkg/ginx"
)

// Get 查询账号接口
// GET /api/v1/accounts/:id
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
