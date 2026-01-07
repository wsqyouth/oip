package etaccount

import (
	"errors"
	"time"
)

// 错误定义
var (
	ErrInvalidAccountID = errors.New("invalid account ID")
	ErrInvalidName      = errors.New("account name cannot be empty")
	ErrInvalidEmail     = errors.New("invalid email format")
)

// Account 账号实体
type Account struct {
	ID        int64     // 账号ID
	Name      string    // 账号名称
	Email     string    // 邮箱
	CreatedAt time.Time // 创建时间
}

// NewAccount 创建账号（工厂方法）
// id: 账号ID，如果为0表示新创建的账号（ID将由数据库自动生成）
func NewAccount(id int64, name, email string) (*Account, error) {
	// 业务规则校验
	if id < 0 {
		return nil, ErrInvalidAccountID
	}
	if name == "" {
		return nil, ErrInvalidName
	}
	if email == "" {
		return nil, ErrInvalidEmail
	}

	return &Account{
		ID:        id,
		Name:      name,
		Email:     email,
		CreatedAt: time.Now(),
	}, nil
}
