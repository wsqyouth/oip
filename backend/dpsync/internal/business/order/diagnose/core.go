package diagnose

import (
	"context"
	"errors"
	"time"

	"oip/dpsync/internal/business/order/diagnose/services"
)

// PreProcess 预处理
func (h *DiagnoseHandler) PreProcess(ctx context.Context) error {
	if h.payload.OrderID == "" {
		return errors.New("order_id is required")
	}

	if h.payload.AccountID <= 0 {
		return errors.New("account_id is invalid")
	}

	return nil
}

// Process 核心处理
func (h *DiagnoseHandler) Process(ctx context.Context) error {
	input := &services.DiagnoseInput{
		RequestID:       h.GetMeta().RequestID,
		OrderID:         h.payload.OrderID,
		AccountID:       h.payload.AccountID,
		MerchantOrderNo: h.payload.MerchantOrderNo,
		Shipment:        h.payload.Shipment,
	}

	compositeHandler := services.NewCompositeHandler()
	result, err := compositeHandler.Diagnose(ctx, input)
	if err != nil {
		return err
	}

	h.diagnosisResult = result

	return nil
}

// PostProcess 后处理
func (h *DiagnoseHandler) PostProcess(ctx context.Context) error {
	err := h.GetResulter().Set(ctx, &DiagnosisResultData{
		Items:       h.diagnosisResult.Items,
		OrderID:     h.payload.OrderID,
		ProcessedAt: time.Now().Unix(),
	})
	if err != nil {
		return err
	}

	output := h.GetResulter().Get(ctx)
	h.SetOutput(output)

	return h.sendCallback(ctx)
}

// sendCallback 发送回调
func (h *DiagnoseHandler) sendCallback(ctx context.Context) error {
	input := &services.DiagnoseInput{
		RequestID:       h.GetMeta().RequestID,
		OrderID:         h.payload.OrderID,
		AccountID:       h.payload.AccountID,
		MerchantOrderNo: h.payload.MerchantOrderNo,
		Shipment:        h.payload.Shipment,
	}

	return h.diagnosisService.ExecuteDiagnosis(ctx, input)
}
