package business

import (
	"context"
	"encoding/json"

	"oip/common/model"
)

// CompositeHandler 复合诊断处理器（组装所有诊断结果）
type CompositeHandler struct {
	shippingCalc   *ShippingCalculator
	anomalyChecker *AnomalyChecker
}

// NewCompositeHandler 创建复合诊断处理器实例
func NewCompositeHandler() *CompositeHandler {
	return &CompositeHandler{
		shippingCalc:   NewShippingCalculator(),
		anomalyChecker: NewAnomalyChecker(),
	}
}

// DiagnoseInput 诊断输入参数（所有数据从 payload 传入）
type DiagnoseInput struct {
	RequestID       string                 `json:"request_id"`
	OrderID         string                 `json:"order_id"`
	AccountID       int64                  `json:"account_id"`
	MerchantOrderNo string                 `json:"merchant_order_no"`
	Shipment        map[string]interface{} `json:"shipment"`
}

// Diagnose 执行完整的订单诊断流程
// 返回 DiagnosisResultData（包含 shipping 和 anomaly 两个诊断项）
func (h *CompositeHandler) Diagnose(ctx context.Context, input *DiagnoseInput) (*model.DiagnosisResultData, error) {
	items := make([]model.DiagnosisItem, 0, 2)

	// 1. 物流费率诊断
	shippingItem := h.diagnoseShipping(ctx, input)
	items = append(items, shippingItem)

	// 2. 异常检测诊断
	anomalyItem := h.diagnoseAnomaly(ctx, input)
	items = append(items, anomalyItem)

	return &model.DiagnosisResultData{
		Items: items,
	}, nil
}

// diagnoseShipping 执行物流费率诊断
func (h *CompositeHandler) diagnoseShipping(ctx context.Context, input *DiagnoseInput) model.DiagnosisItem {
	result, err := h.shippingCalc.Calculate(ctx, input.OrderID, input.Shipment)
	if err != nil {
		return model.DiagnosisItem{
			Type:   model.DiagnosisTypeShipping,
			Status: model.DiagnosisStatusFailed,
			Error:  err.Error(),
		}
	}

	// 序列化结果为 JSON
	dataJSON, err := json.Marshal(result)
	if err != nil {
		return model.DiagnosisItem{
			Type:   model.DiagnosisTypeShipping,
			Status: model.DiagnosisStatusFailed,
			Error:  "Failed to marshal shipping result: " + err.Error(),
		}
	}

	return model.DiagnosisItem{
		Type:     model.DiagnosisTypeShipping,
		Status:   model.DiagnosisStatusSuccess,
		DataJSON: dataJSON,
	}
}

// diagnoseAnomaly 执行异常检测诊断
func (h *CompositeHandler) diagnoseAnomaly(ctx context.Context, input *DiagnoseInput) model.DiagnosisItem {
	result, err := h.anomalyChecker.Check(ctx, input.Shipment)
	if err != nil {
		return model.DiagnosisItem{
			Type:   model.DiagnosisTypeAnomaly,
			Status: model.DiagnosisStatusFailed,
			Error:  err.Error(),
		}
	}

	// 序列化结果为 JSON
	dataJSON, err := json.Marshal(result)
	if err != nil {
		return model.DiagnosisItem{
			Type:   model.DiagnosisTypeAnomaly,
			Status: model.DiagnosisStatusFailed,
			Error:  "Failed to marshal anomaly result: " + err.Error(),
		}
	}

	return model.DiagnosisItem{
		Type:     model.DiagnosisTypeAnomaly,
		Status:   model.DiagnosisStatusSuccess,
		DataJSON: dataJSON,
	}
}
