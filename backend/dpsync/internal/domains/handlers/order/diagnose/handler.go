package diagnose

import (
	"context"
	"encoding/json"
	"fmt"

	"oip/common/model"
	"oip/dpsync/internal/business"
	"oip/dpsync/internal/domains/common"
	"oip/dpsync/internal/domains/common/job"
	"oip/dpsync/internal/domains/common/response"
)

// DiagnoseHandler 订单诊断 Handler
type DiagnoseHandler struct {
	ctx     context.Context
	meta    *job.Meta
	jobData *model.OrderDiagnoseData
}

// NewDiagnoseHandler 创建诊断 Handler
// 解析标准化 Job 消息
func NewDiagnoseHandler(ctx context.Context, meta *job.Meta, payload interface{}) (common.HandlerServ, error) {
	// 解析 payload（业务数据）
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshal payload failed: %w", err)
	}

	var bizData model.OrderDiagnoseBusinessData
	if err := json.Unmarshal(payloadBytes, &bizData); err != nil {
		return nil, fmt.Errorf("unmarshal business data failed: %w", err)
	}

	// 校验必填字段
	if bizData.OrderID == "" {
		return nil, fmt.Errorf("order_id is required")
	}
	if bizData.AccountID == 0 {
		return nil, fmt.Errorf("account_id is required")
	}

	// 包装为完整的 OrderDiagnoseData（兼容原有结构）
	jobData := &model.OrderDiagnoseData{
		RequestID:  meta.RequestID,
		OrgID:      meta.OrgID,
		ActionType: meta.ActionType,
		ID:         meta.ID,
		Data:       bizData,
	}

	return &DiagnoseHandler{
		ctx:     ctx,
		meta:    meta,
		jobData: jobData,
	}, nil
}

// GetProcess 处理诊断请求
func (h *DiagnoseHandler) GetProcess() *response.Response {
	// 创建结果
	result := response.NewDiagnosisResult()

	// 处理业务逻辑
	err := h.process(result)

	// 包装响应
	resp := &response.Response{}
	resp.WrapResponse(result, h.meta, err)

	return resp
}

// process 业务处理逻辑
func (h *DiagnoseHandler) process(result *response.DiagnosisResult) error {
	// 打印开始日志
	logData := map[string]interface{}{
		"handler":    "DiagnoseHandler",
		"action":     "order_diagnose",
		"request_id": h.jobData.RequestID,
		"order_id":   h.jobData.Data.OrderID,
		"account_id": h.jobData.Data.AccountID,
		"phase":      "Full diagnosis with callback queue",
	}

	logJSON, _ := json.MarshalIndent(logData, "", "  ")
	fmt.Printf("\n=== DiagnoseHandler Process ===\n%s\n", string(logJSON))

	// 从 Context 获取 DiagnosisService
	diagnosisService, ok := h.ctx.Value("diagnosis_service").(*business.DiagnosisService)
	if !ok || diagnosisService == nil {
		return fmt.Errorf("DiagnosisService not found in context")
	}

	// 构造诊断输入
	input := &business.DiagnoseInput{
		RequestID:       h.jobData.RequestID,
		OrderID:         h.jobData.Data.OrderID,
		AccountID:       h.jobData.Data.AccountID,
		MerchantOrderNo: h.jobData.Data.MerchantOrderNo,
		Shipment:        h.jobData.Data.Shipment,
	}

	// 调用 DiagnosisService 执行诊断并发送回调
	if err := diagnosisService.ExecuteDiagnosis(h.ctx, input); err != nil {
		fmt.Printf("Error: %v\n==============================\n\n", err)
		return err
	}

	fmt.Printf("Diagnosis completed and callback sent successfully\n")
	fmt.Printf("==============================\n\n")

	return nil
}
