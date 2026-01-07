package business

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"oip/common/model"
	"oip/dpsync/pkg/lmstfy"
)

// DiagnosisService 诊断服务（仅负责诊断逻辑，不涉及 DB 操作）
// 职责：执行诊断 → 发送回调到 callback 队列
type DiagnosisService struct {
	compositeHandler *CompositeHandler
	lmstfyClient     *lmstfy.Client
	callbackQueue    string
}

// NewDiagnosisService 创建诊断服务实例
func NewDiagnosisService(
	lmstfyClient *lmstfy.Client,
	callbackQueue string,
) *DiagnosisService {
	return &DiagnosisService{
		compositeHandler: NewCompositeHandler(),
		lmstfyClient:     lmstfyClient,
		callbackQueue:    callbackQueue,
	}
}

// ExecuteDiagnosis 执行诊断并发送回调
// 返回 error 表示整个流程失败（诊断失败或回调发送失败）
func (s *DiagnosisService) ExecuteDiagnosis(ctx context.Context, input *DiagnoseInput) error {
	// 1. 执行诊断（不查询 DB，使用 payload 传入的数据）
	diagnosisData, diagErr := s.compositeHandler.Diagnose(ctx, input)

	// 2. 构造回调消息
	callback := model.OrderDiagnoseCallback{
		RequestID:   input.RequestID,
		OrderID:     input.OrderID,
		AccountID:   input.AccountID,
		ProcessedAt: time.Now().Unix(),
	}

	if diagErr != nil {
		// 诊断失败
		callback.Status = model.CallbackStatusFailed
		callback.Error = diagErr.Error()
	} else {
		// 诊断成功
		callback.Status = model.CallbackStatusSuccess
		callback.DiagnosisResult = diagnosisData
	}

	// 3. 序列化回调消息为 JSON
	callbackJSON, err := json.Marshal(callback)
	if err != nil {
		return fmt.Errorf("failed to marshal callback: %w", err)
	}

	// 4. 发送回调到 callback 队列
	// ttl=0 表示永不过期, delay=0 表示立即可用
	if err := s.lmstfyClient.Publish(s.callbackQueue, callbackJSON, 0, 0); err != nil {
		return fmt.Errorf("failed to publish callback: %w", err)
	}

	return nil
}
