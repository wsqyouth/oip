package svorder

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"oip/dpmain/internal/app/domains/entity/etorder"
	"oip/dpmain/internal/app/domains/modules/mddiagnosis"
	"oip/dpmain/internal/app/domains/modules/mdorder"
)

// OrderService 订单服务，负责订单业务编排
type OrderService struct {
	orderModule     *mdorder.OrderModule
	diagnosisModule *mddiagnosis.DiagnosisModule
}

// NewOrderService 创建订单服务实例
func NewOrderService(orderModule *mdorder.OrderModule, diagnosisModule *mddiagnosis.DiagnosisModule) *OrderService {
	return &OrderService{
		orderModule:     orderModule,
		diagnosisModule: diagnosisModule,
	}
}

// CreateOrder 创建订单（完整业务流程）
// 1. 验证 account 存在
// 2. 检查订单重复
// 3. 验证货件信息
// 4. 创建订单并落库
// 5. 发布到诊断队列
// 6. Smart Wait（等待诊断结果）
func (s *OrderService) CreateOrder(ctx context.Context, accountID int64, merchantOrderNo string, shipment *etorder.Shipment, waitSeconds int) (*etorder.Order, error) {
	exists, err := s.orderModule.AccountExists(ctx, accountID)
	if err != nil {
		return nil, fmt.Errorf("check account exists failed: %w", err)
	}
	if !exists {
		return nil, errors.New("account not found")
	}

	existing, err := s.orderModule.GetOrderByAccountAndMerchantNo(ctx, accountID, merchantOrderNo)
	if err != nil {
		return nil, fmt.Errorf("check order duplicate failed: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("order already exists: merchant_order_no=%s", merchantOrderNo)
	}

	if err := s.validateShipment(shipment); err != nil {
		return nil, fmt.Errorf("validate shipment failed: %w", err)
	}

	order, err := etorder.NewOrder(uuid.New().String(), accountID, merchantOrderNo, shipment)
	if err != nil {
		return nil, fmt.Errorf("create order entity failed: %w", err)
	}

	if err := s.orderModule.CreateOrder(ctx, order); err != nil {
		return nil, fmt.Errorf("save order failed: %w", err)
	}

	// 5. 发布到诊断队列
	if err := s.diagnosisModule.PublishDiagnoseJob(ctx, order); err != nil {
		// 发布失败只记录日志，不影响订单创建成功
		log.Printf("[WARN] publish diagnose job failed: order_id=%s, error=%v", order.ID, err)
	}

	// 6. Smart Wait（等待诊断结果）
	if waitSeconds > 0 {
		timeout := time.Duration(waitSeconds) * time.Second
		result, err := s.diagnosisModule.WaitForDiagnosisResult(ctx, order.ID, timeout)

		// 修复：正确处理错误
		if err != nil {
			// 超时或订阅失败，只记录日志
			log.Printf("[WARN] wait for diagnosis result failed: order_id=%s, error=%v", order.ID, err)
			return order, nil // 返回订单，状态仍为 DIAGNOSING
		}

		if result != nil {
			// 更新内存中的 Order 实体
			if err := order.UpdateDiagnoseResult(result); err != nil {
				return nil, fmt.Errorf("update order entity failed: %w", err)
			}

			// 持久化到 DB
			if err := s.orderModule.UpdateDiagnoseResult(ctx, order.ID, result); err != nil {
				// 严重问题：内存已更新，DB 更新失败
				log.Printf("[ERROR] persist diagnose result failed: order_id=%s, error=%v", order.ID, err)
				return nil, fmt.Errorf("persist diagnose result failed: %w", err)
			}
		}
	}

	return order, nil
}

// GetOrder 查询订单
func (s *OrderService) GetOrder(ctx context.Context, orderID string) (*etorder.Order, error) {
	return s.orderModule.GetOrder(ctx, orderID)
}

// ListOrders 查询订单列表
func (s *OrderService) ListOrders(ctx context.Context, accountID int64, page, limit int) ([]*etorder.Order, int64, error) {
	return s.orderModule.ListOrders(ctx, accountID, page, limit)
}

// validateShipment 验证货件信息
func (s *OrderService) validateShipment(shipment *etorder.Shipment) error {
	if shipment == nil {
		return errors.New("shipment is required")
	}
	if shipment.ShipFrom == nil || shipment.ShipTo == nil {
		return errors.New("ship_from and ship_to are required")
	}
	if len(shipment.Parcels) == 0 {
		return errors.New("parcels cannot be empty")
	}
	return nil
}
