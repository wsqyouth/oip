package diagnose

import "context"

// DiagnosisResulter 诊断结果处理器
type DiagnosisResulter struct {
	srcData interface{}
	dstData interface{}
}

// NewDiagnosisResulter 创建诊断结果处理器
func NewDiagnosisResulter() *DiagnosisResulter {
	return &DiagnosisResulter{}
}

// Set 设置业务结果数据
func (r *DiagnosisResulter) Set(ctx context.Context, data interface{}) error {
	r.srcData = data

	resultData := data.(*DiagnosisResultData)

	r.dstData = &DiagnosisOutput{
		Items:       resultData.Items,
		OrderID:     resultData.OrderID,
		ProcessedAt: resultData.ProcessedAt,
	}

	return nil
}

// Get 获取格式化后的输出
func (r *DiagnosisResulter) Get(ctx context.Context) interface{} {
	return r.dstData
}
