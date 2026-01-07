package response

import (
	"oip/dpsync/internal/domains/common/job"
	"oip/dpsync/pkg/errorutil"
)

// DiagnosisResult 诊断结果（实现 ResultI 接口）
// 直接复用 common/model/diagnosis_result.go 的结构
type DiagnosisResult struct {
	ID     string           `json:"id"`
	Status string           `json:"status"`
	Data   interface{}      `json:"data"`
	Error  *errorutil.Error `json:"error,omitempty"`
}

const (
	DiagnosisStatusSuccess = "SUCCESS"
	DiagnosisStatusFailed  = "FAILED"
)

// NewDiagnosisResult 创建诊断结果
func NewDiagnosisResult() *DiagnosisResult {
	return &DiagnosisResult{}
}

// Set 实现 ResultI 接口
func (r *DiagnosisResult) Set(meta *job.Meta, err error) {
	r.ID = meta.ID
	if err != nil {
		r.Status = DiagnosisStatusFailed
		r.Error = errorutil.Wrap(err)
	} else {
		r.Status = DiagnosisStatusSuccess
	}
}

// GetStatus 实现 ResultI 接口
func (r *DiagnosisResult) GetStatus() string {
	return r.Status
}
