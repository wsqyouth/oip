package rporder

import (
	"context"

	"oip/common/model"
	"oip/dpmain/internal/app/domains/entity/etorder"
)

// OrderRepository 订单仓储接口（只定义，不实现）
// 实现在 infra/persistence 层
type OrderRepository interface {
	// Create 创建订单
	Create(ctx context.Context, order *etorder.Order) error

	// GetByID 根据ID查询订单
	GetByID(ctx context.Context, orderID string) (*etorder.Order, error)

	// GetByAccountAndMerchantNo 根据账号ID和商家订单号查询
	GetByAccountAndMerchantNo(ctx context.Context, accountID int64, merchantOrderNo string) (*etorder.Order, error)

	// UpdateDiagnoseResult 更新诊断结果（旧方法，保持兼容）
	UpdateDiagnoseResult(ctx context.Context, orderID string, result *etorder.DiagnoseResult) error

	// UpdateStatus 更新订单状态
	UpdateStatus(ctx context.Context, orderID string, status etorder.OrderStatus) error

	// UpdateDiagnosisResult 更新诊断结果（新方法，支持成功/失败两种情况）
	// diagnosisResult: 诊断结果（成功时传入，失败时传 nil）
	// status: 订单状态（DIAGNOSED 或 FAILED）
	// errorMsg: 错误信息（失败时传入）
	UpdateDiagnosisResult(ctx context.Context, orderID string, diagnosisResult *model.DiagnosisResultData, status string, errorMsg string) error

	// List 查询订单列表
	List(ctx context.Context, accountID int64, page, limit int) ([]*etorder.Order, int64, error)
}
