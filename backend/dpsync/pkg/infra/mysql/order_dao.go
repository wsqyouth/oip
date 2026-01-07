package mysql

import (
	"context"
	"encoding/json"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"oip/common/entity"
	"oip/common/model"
)

// OrderDAO 订单数据访问对象
type OrderDAO struct {
	db *gorm.DB
}

// NewOrderDAO 创建 OrderDAO 实例
func NewOrderDAO(dsn string) (*OrderDAO, error) {
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &OrderDAO{
		db: db,
	}, nil
}

// UpdateDiagnosisResult 更新订单的诊断结果
// 参数：
//   - ctx: 上下文
//   - orderID: 订单 ID
//   - result: 诊断结果数据
//   - status: 订单状态（DIAGNOSED/FAILED）
//   - errorMsg: 错误消息（失败时）
func (dao *OrderDAO) UpdateDiagnosisResult(
	ctx context.Context,
	orderID string,
	result *model.DiagnosisResultData,
	status string,
	errorMsg string,
) error {
	// 序列化诊断结果为 JSON
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal diagnosis result: %w", err)
	}

	// 构造更新字段
	updates := map[string]interface{}{
		"status":          status,
		"diagnose_result": resultJSON,
	}

	if errorMsg != "" {
		updates["error_message"] = errorMsg
	}

	// 执行更新
	dbResult := dao.db.WithContext(ctx).
		Model(&entity.Order{}).
		Where("id = ?", orderID).
		Updates(updates)

	if dbResult.Error != nil {
		return fmt.Errorf("failed to update order: %w", dbResult.Error)
	}

	if dbResult.RowsAffected == 0 {
		return fmt.Errorf("order not found: %s", orderID)
	}

	return nil
}

// GetOrderByID 根据订单 ID 获取订单
func (dao *OrderDAO) GetOrderByID(ctx context.Context, orderID string) (*entity.Order, error) {
	var order entity.Order
	result := dao.db.WithContext(ctx).Where("id = ?", orderID).First(&order)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get order: %w", result.Error)
	}
	return &order, nil
}

// Close 关闭数据库连接
func (dao *OrderDAO) Close() error {
	sqlDB, err := dao.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
