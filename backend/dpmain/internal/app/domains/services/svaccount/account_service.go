package svaccount

import (
	"context"
	"errors"
	"fmt"

	"oip/dpmain/internal/app/domains/entity/etaccount"
	"oip/dpmain/internal/app/domains/modules/mdaccount"
	"oip/dpmain/internal/app/pkg/idgen"
)

// AccountService 账号服务，负责账号业务编排
type AccountService struct {
	accountModule *mdaccount.AccountModule
}

// NewAccountService 创建账号服务实例
func NewAccountService(accountModule *mdaccount.AccountModule) *AccountService {
	return &AccountService{
		accountModule: accountModule,
	}
}

// CreateAccount 创建账号（完整业务流程）
// 1. 检查邮箱是否重复
// 2. 生成分布式ID
// 3. 创建账号并落库
func (s *AccountService) CreateAccount(ctx context.Context, name, email string) (*etaccount.Account, error) {
	existing, err := s.accountModule.GetAccountByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("check email duplicate failed: %w", err)
	}
	if existing != nil {
		return nil, errors.New("email already exists")
	}

	// 生成分布式ID（例如: 25610, 95918）
	accountID := idgen.GenerateID()

	account, err := etaccount.NewAccount(accountID, name, email)
	if err != nil {
		return nil, fmt.Errorf("create account entity failed: %w", err)
	}

	if err := s.accountModule.CreateAccount(ctx, account); err != nil {
		return nil, fmt.Errorf("save account failed: %w", err)
	}

	return account, nil
}

// GetAccount 查询账号
func (s *AccountService) GetAccount(ctx context.Context, accountID int64) (*etaccount.Account, error) {
	return s.accountModule.GetAccount(ctx, accountID)
}
