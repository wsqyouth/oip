package rpaccount

import (
	"context"

	"oip/dpmain/internal/app/domains/entity/etaccount"
)

// AccountRepository 账号仓储接口
type AccountRepository interface {
	// Create 创建账号
	Create(ctx context.Context, account *etaccount.Account) error

	// GetByID 根据ID查询账号
	GetByID(ctx context.Context, accountID int64) (*etaccount.Account, error)

	// GetByEmail 根据邮箱查询账号
	GetByEmail(ctx context.Context, email string) (*etaccount.Account, error)

	// Exists 检查账号是否存在
	Exists(ctx context.Context, accountID int64) (bool, error)
}
