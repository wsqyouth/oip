package entity

import "time"

// Account 账号实体
type Account struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement"`
	Name      string    `gorm:"column:name;type:varchar(255);not null"`
	Email     string    `gorm:"column:email;type:varchar(255);uniqueIndex:uk_email;not null"`
	CreatedAt time.Time `gorm:"column:created_at;not null"`
}

// TableName 指定表名
func (Account) TableName() string {
	return "accounts"
}
