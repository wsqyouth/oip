package mdaccount

import (
	"context"

	"oip/dpmain/internal/app/domains/entity/etaccount"
	"oip/dpmain/internal/app/domains/repo/rpaccount"
)

// AccountModule 账号模块
type AccountModule struct {
	accountRepo rpaccount.AccountRepository
}

// NewAccountModule 创建账号模块
func NewAccountModule(accountRepo rpaccount.AccountRepository) *AccountModule {
	return &AccountModule{
		accountRepo: accountRepo,
	}
}

// CreateAccount 创建账号（数据操作）
func (m *AccountModule) CreateAccount(ctx context.Context, account *etaccount.Account) error {
	return m.accountRepo.Create(ctx, account)
}

// GetAccount 查询账号
func (m *AccountModule) GetAccount(ctx context.Context, accountID int64) (*etaccount.Account, error) {
	return m.accountRepo.GetByID(ctx, accountID)
}

// GetAccountByEmail 根据邮箱查询账号（检查重复）
func (m *AccountModule) GetAccountByEmail(ctx context.Context, email string) (*etaccount.Account, error) {
	return m.accountRepo.GetByEmail(ctx, email)
}
