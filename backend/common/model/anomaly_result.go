package model

// AnomalyResult 异常检测结果
type AnomalyResult struct {
	HasRisk bool          `json:"has_risk"`
	Issues  []AnomalyItem `json:"issues"`
}

// AnomalyItem 单个异常项
type AnomalyItem struct {
	Type    string `json:"type"`    // HIGH_VALUE/REMOTE_AREA/SKU_MISSING
	Level   string `json:"level"`   // INFO/WARNING/CRITICAL
	Message string `json:"message"` // 人类可读描述
}

// 异常级别常量
const (
	AnomalyLevelInfo     = "INFO"
	AnomalyLevelWarning  = "WARNING"
	AnomalyLevelCritical = "CRITICAL"
)

// 异常类型常量
const (
	AnomalyTypeHighValue    = "HIGH_VALUE"
	AnomalyTypeHeavyPackage = "HEAVY_PACKAGE"
	AnomalyTypeSKUMissing   = "SKU_MISSING"
)
