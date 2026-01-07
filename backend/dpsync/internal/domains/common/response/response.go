package response

import (
	"oip/dpsync/internal/domains/common/job"
	"oip/dpsync/pkg/errorutil"
)

// ResultI 业务结果接口
type ResultI interface {
	// Set 设置元数据和错误
	Set(meta *job.Meta, err error)

	// GetStatus 获取状态
	GetStatus() string
}

// Response 统一响应结构
type Response struct {
	Error     *errorutil.Error `json:"error"`
	Result    ResultI          `json:"result"`
	Processed bool             `json:"processed"`
	Meta      interface{}      `json:"meta"`
}

// WrapResponse 包装响应
func (r *Response) WrapResponse(result ResultI, meta *job.Meta, err error) {
	result.Set(meta, err)

	if err == nil {
		r.Processed = true
	}
	r.Meta = meta
	r.Error = errorutil.UnWrapResponse(err)
	r.Result = result
}

// WrapResults 包装成数组（用于序列化）
func (r *Response) WrapResults() []interface{} {
	if r.Error != nil {
		return []interface{}{r.Error, r.Result, r.Processed, r.Meta}
	}
	return []interface{}{r.Error, r.Result, r.Processed}
}
