package model

import "encoding/json"

// DiagnosisResultData 诊断结果容器
type DiagnosisResultData struct {
	Items []DiagnosisItem `json:"items"`
}

// DiagnosisItem 单个诊断项
type DiagnosisItem struct {
	Type     string          `json:"type"`      // shipping/anomaly/risk/compliance
	Status   string          `json:"status"`    // SUCCESS/FAILED
	DataJSON json.RawMessage `json:"data_json"` // 具体数据
	Error    string          `json:"error,omitempty"`
}

// 诊断项状态常量
const (
	DiagnosisStatusSuccess = "SUCCESS"
	DiagnosisStatusFailed  = "FAILED"
)

// 诊断类型常量
const (
	DiagnosisTypeShipping = "shipping"
	DiagnosisTypeAnomaly  = "anomaly"
)
