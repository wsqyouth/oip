package framework

import (
	"context"
	"encoding/json"
	"fmt"
)

// BaseHandler 抽象基类
// 提供基础设施方法，不包含业务流程控制
type BaseHandler struct {
	meta       *JobMeta    // Job 元信息
	rawData    []byte      // 原始 Job 数据（Lmstfy 消息原始 bytes）
	bizPayload interface{} // 业务数据（job.Payload.Data.Data 部分）
	output     interface{} // 最终输出结果
	resulter   Resulter    // 结果处理器（业务提供）
}

// Job 标准 Job 结构
type Job struct {
	Payload *JobPayload `json:"payload"`
}

type JobPayload struct {
	Data *JobPayloadData `json:"data"`
}

type JobPayloadData struct {
	RequestID  string      `json:"request_id"`
	ActionType string      `json:"action_type"`
	OrgID      string      `json:"org_id"`
	ID         string      `json:"id"`
	Data       interface{} `json:"data"`
}

// JobMeta Job 元信息
type JobMeta struct {
	RequestID  string
	ActionType string
	OrgID      string
	ID         string
}

// Response 标准响应结构
type Response struct {
	Error     interface{} `json:"error"`
	Result    interface{} `json:"result"`
	Processed bool        `json:"processed"`
	Meta      *JobMeta    `json:"meta,omitempty"`
}

// ParseJob 解析 lmstfy Job 标准结构
// 将解析后的数据存储到 BaseHandler 成员变量中
func (b *BaseHandler) ParseJob(ctx context.Context, rawData []byte) error {
	b.rawData = rawData

	var job Job
	if err := json.Unmarshal(rawData, &job); err != nil {
		return b.WrapError(err, "unmarshal job failed")
	}

	if job.Payload == nil || job.Payload.Data == nil {
		return b.WrapError(nil, "invalid job structure")
	}

	data := job.Payload.Data
	b.meta = &JobMeta{
		RequestID:  data.RequestID,
		ActionType: data.ActionType,
		OrgID:      data.OrgID,
		ID:         data.ID,
	}

	b.bizPayload = data.Data

	return nil
}

// WrapResponse 包装标准响应
func (b *BaseHandler) WrapResponse(ctx context.Context, output interface{}) ([]byte, error) {
	resp := &Response{
		Error:     nil,
		Result:    output,
		Processed: true,
		Meta:      b.meta,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		return nil, b.WrapError(err, "marshal response failed")
	}

	return data, nil
}

// WrapErrorResponse 包装错误响应
func (b *BaseHandler) WrapErrorResponse(ctx context.Context, err error) ([]byte, error) {
	resp := &Response{
		Error:     err.Error(),
		Result:    nil,
		Processed: false,
		Meta:      b.meta,
	}

	data, marshalErr := json.Marshal(resp)
	if marshalErr != nil {
		return nil, b.WrapError(marshalErr, "marshal error response failed")
	}

	return data, nil
}

// WrapError 统一包装错误
func (b *BaseHandler) WrapError(err error, msg string) error {
	if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}
	return fmt.Errorf("%s", msg)
}

// GetMeta 获取 meta
func (b *BaseHandler) GetMeta() *JobMeta {
	return b.meta
}

// GetRawData 获取原始数据
func (b *BaseHandler) GetRawData() []byte {
	return b.rawData
}

// GetBizPayload 获取业务数据
func (b *BaseHandler) GetBizPayload() interface{} {
	return b.bizPayload
}

// SetOutput 设置输出
func (b *BaseHandler) SetOutput(output interface{}) {
	b.output = output
}

// GetOutput 获取输出
func (b *BaseHandler) GetOutput() interface{} {
	return b.output
}

// SetResulter 设置结果处理器
func (b *BaseHandler) SetResulter(resulter Resulter) {
	b.resulter = resulter
}

// GetResulter 获取结果处理器
func (b *BaseHandler) GetResulter() Resulter {
	return b.resulter
}
