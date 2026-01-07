package business

import (
	"context"
	"fmt"

	"oip/common/model"
)

// AnomalyChecker 异常检测器（规则引擎）
type AnomalyChecker struct{}

// NewAnomalyChecker 创建异常检测器实例
func NewAnomalyChecker() *AnomalyChecker {
	return &AnomalyChecker{}
}

// Check 执行异常检测（基于固定规则）
// 参数 shipment 为物流信息，包含 ship_from、ship_to、parcels 等
func (c *AnomalyChecker) Check(ctx context.Context, shipment map[string]interface{}) (*model.AnomalyResult, error) {
	issues := make([]model.AnomalyItem, 0)

	// 规则 1：检查 parcels 是否存在
	parcels, ok := shipment["parcels"].([]interface{})
	if !ok || len(parcels) == 0 {
		issues = append(issues, model.AnomalyItem{
			Type:    "MISSING_PARCELS",
			Level:   "CRITICAL",
			Message: "Parcels information is missing",
		})

		return &model.AnomalyResult{
			HasRisk: len(issues) > 0,
			Issues:  issues,
		}, nil
	}

	// 规则 2：检查重量异常（TBC: 简化版本,仅作示例）
	// 实际应该解析 parcel.weight 字段
	totalWeight := 0.0
	for _, parcel := range parcels {
		if parcelMap, ok := parcel.(map[string]interface{}); ok {
			if weight, ok := parcelMap["weight"].(map[string]interface{}); ok {
				if value, ok := weight["value"].(float64); ok {
					totalWeight += value
				}
			}
		}
	}

	if totalWeight > 10.0 {
		issues = append(issues, model.AnomalyItem{
			Type:    "HEAVY_PACKAGE",
			Level:   "WARNING",
			Message: fmt.Sprintf("Total weight %.2f kg may incur additional fees", totalWeight),
		})
	}

	// 规则 3：检查 SKU 缺失（TBC: 简化版本）
	for i, parcel := range parcels {
		if parcelMap, ok := parcel.(map[string]interface{}); ok {
			if items, ok := parcelMap["items"].([]interface{}); ok {
				for j, item := range items {
					if itemMap, ok := item.(map[string]interface{}); ok {
						sku, hasSKU := itemMap["sku"].(string)
						if !hasSKU || sku == "" {
							issues = append(issues, model.AnomalyItem{
								Type:    "SKU_MISSING",
								Level:   "CRITICAL",
								Message: fmt.Sprintf("Item #%d in parcel #%d missing SKU", j+1, i+1),
							})
						}
					}
				}
			}
		}
	}

	return &model.AnomalyResult{
		HasRisk: len(issues) > 0,
		Issues:  issues,
	}, nil
}
