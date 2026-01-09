package diagnose

import (
	"context"
	"encoding/json"

	"oip/common/model"
	"oip/dpsync/internal/business/order/diagnose/services"
	"oip/dpsync/internal/framework"
	"oip/dpsync/pkg/lmstfy"
)

// DiagnoseHandler 诊断处理器
type DiagnoseHandler struct {
	framework.BaseHandler

	payload          *DiagnosePayload
	diagnosisService *services.DiagnosisService
	diagnosisResult  *model.DiagnosisResultData
}

// NewDiagnoseHandler 创建诊断处理器
func NewDiagnoseHandler(
	ctx context.Context,
	baseHandler *framework.BaseHandler,
	lmstfyClient *lmstfy.Client,
	callbackQueue string,
) (framework.BusinessHandler, error) {
	bizPayload := baseHandler.GetBizPayload()

	payloadBytes, err := json.Marshal(bizPayload)
	if err != nil {
		return nil, err
	}

	var payload DiagnosePayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, err
	}

	handler := &DiagnoseHandler{
		BaseHandler:      *baseHandler,
		payload:          &payload,
		diagnosisService: services.NewDiagnosisService(lmstfyClient, callbackQueue),
	}

	handler.SetResulter(NewDiagnosisResulter())

	return handler, nil
}

// Handle 处理入口
func (h *DiagnoseHandler) Handle(ctx context.Context) ([]byte, error) {
	processFuncs := []framework.ProcessorFunc{
		h.PreProcess,
		h.Process,
		h.PostProcess,
	}

	preProcessor := framework.NewPreProcessor(processFuncs)
	if err := preProcessor.Run(ctx); err != nil {
		return h.WrapErrorResponse(ctx, err)
	}

	output := h.GetOutput()
	return h.WrapResponse(ctx, output)
}
