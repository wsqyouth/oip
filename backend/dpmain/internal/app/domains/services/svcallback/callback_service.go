package svcallback

import (
	"context"
	"encoding/json"
	"fmt"

	"oip/common/entity"
	"oip/common/model"
	"oip/dpmain/internal/app/domains/repo/rporder"
	"oip/dpmain/internal/app/infra/persistence/redis"
	"oip/dpmain/internal/app/pkg/logger"
)

// CallbackService 回调处理服务
// 职责：
// 1. 处理 dpsync 发送的诊断回调
// 2. 更新 DB 订单状态
// 3. 发送 Redis PubSub 通知（Smart Wait）
type CallbackService struct {
	orderRepo   rporder.OrderRepository
	redisClient *redis.PubSubClient
	logger      logger.Logger
}

// NewCallbackService 创建回调服务实例
func NewCallbackService(
	orderRepo rporder.OrderRepository,
	redisClient *redis.PubSubClient,
	logger logger.Logger,
) *CallbackService {
	return &CallbackService{
		orderRepo:   orderRepo,
		redisClient: redisClient,
		logger:      logger,
	}
}

// HandleCallback 处理诊断回调
// 返回 error 表示处理失败（需要重试）
func (s *CallbackService) HandleCallback(ctx context.Context, callback *model.OrderDiagnoseCallback) error {
	s.logger.InfoContext(ctx, "Processing callback",
		"order_id", callback.OrderID,
		"status", callback.Status,
		"request_id", callback.RequestID,
	)

	// 1. 根据回调状态更新 DB
	if err := s.updateOrderStatus(ctx, callback); err != nil {
		s.logger.ErrorContext(ctx, "Failed to update order status",
			"order_id", callback.OrderID,
			"error", err,
		)
		return fmt.Errorf("update order status failed: %w", err)
	}

	// 2. 发送 Redis PubSub 通知（用于 Smart Wait）
	if err := s.publishNotification(ctx, callback); err != nil {
		// 通知失败不影响整体流程（DB 已更新成功）
		// 只记录日志，不返回错误
		s.logger.WarnContext(ctx, "Failed to publish Redis notification",
			"order_id", callback.OrderID,
			"error", err,
		)
	}

	s.logger.InfoContext(ctx, "Callback processed successfully",
		"order_id", callback.OrderID,
	)

	return nil
}

// updateOrderStatus 根据回调状态更新订单
func (s *CallbackService) updateOrderStatus(ctx context.Context, callback *model.OrderDiagnoseCallback) error {
	if callback.Status == model.CallbackStatusSuccess {
		// 诊断成功：更新状态为 DIAGNOSED，保存诊断结果
		return s.orderRepo.UpdateDiagnosisResult(
			ctx,
			callback.OrderID,
			callback.DiagnosisResult,
			entity.OrderStatusDiagnosed,
			"",
		)
	} else {
		// 诊断失败：更新状态为 FAILED，保存错误信息
		return s.orderRepo.UpdateDiagnosisResult(
			ctx,
			callback.OrderID,
			nil,
			entity.OrderStatusFailed,
			callback.Error,
		)
	}
}

// publishNotification 发送 Redis PubSub 通知（使用订单独立频道）
func (s *CallbackService) publishNotification(ctx context.Context, callback *model.OrderDiagnoseCallback) error {
	// 构造独立频道名称
	channel := fmt.Sprintf("diagnosis:result:%s", callback.OrderID)

	// 构造通知数据（与 dpmain API 期望格式一致）
	var notificationData interface{}
	if callback.Status == model.CallbackStatusSuccess && callback.DiagnosisResult != nil {
		// 成功：发送诊断结果
		notificationData = map[string]interface{}{
			"items": callback.DiagnosisResult.Items,
		}
	} else {
		// 失败：发送错误信息
		notificationData = map[string]interface{}{
			"status": callback.Status,
			"error":  callback.Error,
		}
	}

	// 序列化为 JSON
	payload, err := json.Marshal(notificationData)
	if err != nil {
		return fmt.Errorf("marshal notification failed: %w", err)
	}

	// 发送到 Redis
	if err := s.redisClient.Publish(ctx, channel, string(payload)); err != nil {
		return fmt.Errorf("publish to redis failed: %w", err)
	}

	s.logger.InfoContext(ctx, "Redis notification sent",
		"order_id", callback.OrderID,
		"channel", channel,
	)

	return nil
}
