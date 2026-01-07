package mdorder

import (
	"context"

	"oip/dpmain/internal/app/domains/entity/etorder"
	"oip/dpmain/internal/app/domains/repo/rpaccount"
	"oip/dpmain/internal/app/domains/repo/rporder"
)

// OrderModule 订单模块（业务编排层）
type OrderModule struct {
	orderRepo   rporder.OrderRepository
	accountRepo rpaccount.AccountRepository
}

// NewOrderModule 创建订单模块
func NewOrderModule(
	orderRepo rporder.OrderRepository,
	accountRepo rpaccount.AccountRepository,
) *OrderModule {
	return &OrderModule{
		orderRepo:   orderRepo,
		accountRepo: accountRepo,
	}
}

// CreateOrder 创建订单（数据操作）
func (m *OrderModule) CreateOrder(ctx context.Context, order *etorder.Order) error {
	return m.orderRepo.Create(ctx, order)
}

// GetOrder 查询订单
func (m *OrderModule) GetOrder(ctx context.Context, orderID string) (*etorder.Order, error) {
	return m.orderRepo.GetByID(ctx, orderID)
}

// GetOrderByAccountAndMerchantNo 根据账号ID和商户订单号查询（检查重复）
func (m *OrderModule) GetOrderByAccountAndMerchantNo(ctx context.Context, accountID int64, merchantOrderNo string) (*etorder.Order, error) {
	return m.orderRepo.GetByAccountAndMerchantNo(ctx, accountID, merchantOrderNo)
}

// UpdateDiagnoseResult 更新诊断结果
func (m *OrderModule) UpdateDiagnoseResult(ctx context.Context, orderID string, result *etorder.DiagnoseResult) error {
	return m.orderRepo.UpdateDiagnoseResult(ctx, orderID, result)
}

// ListOrders 查询订单列表
func (m *OrderModule) ListOrders(ctx context.Context, accountID int64, page, limit int) ([]*etorder.Order, int64, error) {
	return m.orderRepo.List(ctx, accountID, page, limit)
}

// AccountExists 检查账号是否存在
func (m *OrderModule) AccountExists(ctx context.Context, accountID int64) (bool, error) {
	return m.accountRepo.Exists(ctx, accountID)
}
