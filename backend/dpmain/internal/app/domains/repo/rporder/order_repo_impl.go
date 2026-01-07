package rporder

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"oip/common/entity"
	"oip/common/model"
	"oip/dpmain/internal/app/domains/entity/etorder"

	"gorm.io/gorm"
)

// OrderRepositoryImpl 订单仓储实现（MySQL）
type OrderRepositoryImpl struct {
	db *gorm.DB
}

// NewOrderRepository 创建订单仓储实例
func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &OrderRepositoryImpl{db: db}
}

// Create 创建订单，将领域对象转换为 GORM 模型后存储
func (r *OrderRepositoryImpl) Create(ctx context.Context, order *etorder.Order) error {
	po, err := r.toGormModel(order)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(po).Error
}

// GetByID 根据ID查询订单，将 GORM 模型转换为领域对象
func (r *OrderRepositoryImpl) GetByID(ctx context.Context, orderID string) (*etorder.Order, error) {
	var po entity.Order
	err := r.db.WithContext(ctx).Where("id = ?", orderID).First(&po).Error
	if err != nil {
		return nil, err
	}
	return r.toDomainModel(&po)
}

// GetByAccountAndMerchantNo 根据账号ID和商户订单号查询（用于检查重复）
func (r *OrderRepositoryImpl) GetByAccountAndMerchantNo(ctx context.Context, accountID int64, merchantOrderNo string) (*etorder.Order, error) {
	var po entity.Order
	err := r.db.WithContext(ctx).
		Where("account_id = ? AND merchant_order_no = ?", accountID, merchantOrderNo).
		First(&po).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return r.toDomainModel(&po)
}

// UpdateDiagnoseResult 更新订单的诊断结果
func (r *OrderRepositoryImpl) UpdateDiagnoseResult(ctx context.Context, orderID string, result *etorder.DiagnoseResult) error {
	resultJSON, err := json.Marshal(result)
	if err != nil {
		return err
	}
	return r.db.WithContext(ctx).
		Model(&entity.Order{}).
		Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"diagnose_result": resultJSON,
			"status":          string(etorder.OrderStatusDiagnosed),
			"updated_at":      time.Now(),
		}).Error
}

// UpdateStatus 更新订单状态
func (r *OrderRepositoryImpl) UpdateStatus(ctx context.Context, orderID string, status etorder.OrderStatus) error {
	return r.db.WithContext(ctx).
		Model(&entity.Order{}).
		Where("id = ?", orderID).
		Updates(map[string]interface{}{
			"status":     string(status),
			"updated_at": time.Now(),
		}).Error
}

// UpdateDiagnosisResult 更新诊断结果（新方法，支持成功/失败两种情况）
func (r *OrderRepositoryImpl) UpdateDiagnosisResult(ctx context.Context, orderID string, diagnosisResult *model.DiagnosisResultData, status string, errorMsg string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	// 成功时保存诊断结果
	if diagnosisResult != nil {
		resultJSON, err := json.Marshal(diagnosisResult)
		if err != nil {
			return err
		}
		updates["diagnose_result"] = resultJSON
	}

	// 失败时保存错误信息（TBC: 当前 entity.Order 没有 error_message 字段）
	// 可以考虑扩展数据库表添加该字段
	// if errorMsg != "" {
	//     updates["error_message"] = errorMsg
	// }

	return r.db.WithContext(ctx).
		Model(&entity.Order{}).
		Where("id = ?", orderID).
		Updates(updates).Error
}

// List 分页查询订单列表
func (r *OrderRepositoryImpl) List(ctx context.Context, accountID int64, page, limit int) ([]*etorder.Order, int64, error) {
	var total int64
	var pos []entity.Order

	query := r.db.WithContext(ctx).Model(&entity.Order{})
	if accountID > 0 {
		query = query.Where("account_id = ?", accountID)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&pos).Error; err != nil {
		return nil, 0, err
	}

	orders := make([]*etorder.Order, 0, len(pos))
	for i := range pos {
		order, err := r.toDomainModel(&pos[i])
		if err != nil {
			return nil, 0, err
		}
		orders = append(orders, order)
	}

	return orders, total, nil
}

// toGormModel 领域对象转换为 GORM 模型
func (r *OrderRepositoryImpl) toGormModel(order *etorder.Order) (*entity.Order, error) {
	shipmentJSON, err := json.Marshal(order.Shipment)
	if err != nil {
		return nil, err
	}

	po := &entity.Order{
		ID:              order.ID,
		AccountID:       order.AccountID,
		MerchantOrderNo: order.MerchantOrderNo,
		RawData:         shipmentJSON,
		Status:          string(order.Status),
		CreatedAt:       order.CreatedAt,
		UpdatedAt:       order.UpdatedAt,
	}

	if order.DiagnoseResult != nil {
		resultJSON, err := json.Marshal(order.DiagnoseResult)
		if err != nil {
			return nil, err
		}
		po.DiagnoseResult = resultJSON
	}

	return po, nil
}

// toDomainModel GORM 模型转换为领域对象
func (r *OrderRepositoryImpl) toDomainModel(po *entity.Order) (*etorder.Order, error) {
	var shipment etorder.Shipment
	if err := json.Unmarshal(po.RawData, &shipment); err != nil {
		return nil, err
	}

	order := &etorder.Order{
		ID:              po.ID,
		AccountID:       po.AccountID,
		MerchantOrderNo: po.MerchantOrderNo,
		Shipment:        &shipment,
		Status:          etorder.OrderStatus(po.Status),
		CreatedAt:       po.CreatedAt,
		UpdatedAt:       po.UpdatedAt,
	}

	if len(po.DiagnoseResult) > 0 {
		var result etorder.DiagnoseResult
		if err := json.Unmarshal(po.DiagnoseResult, &result); err != nil {
			return nil, err
		}
		order.DiagnoseResult = &result
	}

	return order, nil
}
