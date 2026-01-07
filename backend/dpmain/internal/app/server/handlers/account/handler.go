package account

import "oip/dpmain/internal/app/domains/services/svaccount"

// AccountHandler 账号 HTTP 处理器
type AccountHandler struct {
	accountService *svaccount.AccountService
}

// NewAccountHandler 创建账号处理器实例
func NewAccountHandler(accountService *svaccount.AccountService) *AccountHandler {
	return &AccountHandler{
		accountService: accountService,
	}
}
