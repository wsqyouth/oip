package entity

import (
	"time"

	"gorm.io/datatypes"
)

// Order 订单实体（包含诊断结果）
type Order struct {
	// 基础字段
	ID              string `gorm:"column:id;primaryKey;type:varchar(64)"`
	AccountID       int64  `gorm:"column:account_id;not null;index:idx_account_status"`
	MerchantOrderNo string `gorm:"column:merchant_order_no;type:varchar(128);not null;uniqueIndex:uk_account_merchant"`

	// 订单数据
	RawData datatypes.JSON `gorm:"column:shipment;type:json;not null"`

	// 诊断状态与结果
	Status         string         `gorm:"column:status;type:varchar(16);not null;default:'DIAGNOSING';index:idx_account_status"`
	DiagnoseResult datatypes.JSON `gorm:"column:diagnose_result;type:json"`

	// 时间戳
	CreatedAt time.Time `gorm:"column:created_at;not null;index:idx_created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null"`
}

// TableName 指定表名
func (Order) TableName() string {
	return "orders"
}

// 订单状态常量
const (
	OrderStatusDiagnosing = "DIAGNOSING"
	OrderStatusDiagnosed  = "DIAGNOSED"
	OrderStatusFailed     = "FAILED"
)
