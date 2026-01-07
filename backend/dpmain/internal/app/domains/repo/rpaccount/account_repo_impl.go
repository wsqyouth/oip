package rpaccount

import (
	"context"

	"gorm.io/gorm"
	"oip/common/entity"
	"oip/dpmain/internal/app/domains/entity/etaccount"
)

// AccountRepositoryImpl 账号仓储实现（MySQL）
type AccountRepositoryImpl struct {
	db *gorm.DB
}

// NewAccountRepository 创建账号仓储实例
func NewAccountRepository(db *gorm.DB) AccountRepository {
	return &AccountRepositoryImpl{db: db}
}

// Create 创建账号
func (r *AccountRepositoryImpl) Create(ctx context.Context, account *etaccount.Account) error {
	po := &entity.Account{
		ID:    account.ID,
		Name:  account.Name,
		Email: account.Email,
	}
	if err := r.db.WithContext(ctx).Create(po).Error; err != nil {
		return err
	}
	// 将数据库生成的ID回写到领域对象
	account.ID = po.ID
	return nil
}

// GetByID 根据ID查询账号
func (r *AccountRepositoryImpl) GetByID(ctx context.Context, accountID int64) (*etaccount.Account, error) {
	var po entity.Account
	err := r.db.WithContext(ctx).Where("id = ?", accountID).First(&po).Error
	if err != nil {
		return nil, err
	}

	// 转换为领域对象
	return etaccount.NewAccount(po.ID, po.Name, po.Email)
}

// GetByEmail 根据邮箱查询账号（用于检查重复）
func (r *AccountRepositoryImpl) GetByEmail(ctx context.Context, email string) (*etaccount.Account, error) {
	var po entity.Account
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&po).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return etaccount.NewAccount(po.ID, po.Name, po.Email)
}

// Exists 检查账号是否存在
func (r *AccountRepositoryImpl) Exists(ctx context.Context, accountID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entity.Account{}).Where("id = ?", accountID).Count(&count).Error
	return count > 0, err
}
