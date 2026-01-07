package lmstfyx

import (
	"context"

	"github.com/bitleak/lmstfy/client"
)

// Proc 业务处理函数类型（GetProcess 的函数签名）
// 参数：ctx 上下文，job 原始 lmstfy Job
// 返回：JobResp 处理结果
type Proc func(ctx context.Context, job *client.Job) *JobResp

// JobRespStatus 消息处理结果状态
type JobRespStatus int

const (
	// JobRespStatusSuccess 处理成功，ACK 消息
	JobRespStatusSuccess JobRespStatus = iota
	// JobRespStatusRelease 需要重试，Release 消息（延迟重新投递）
	JobRespStatusRelease
	// JobRespStatusBury 处理失败且不可重试，Bury 消息（移到死信队列）
	JobRespStatusBury
)

// JobResp 消息处理结果
type JobResp struct {
	Action JobRespStatus // 处理动作
	Data   []byte        // 响应数据（可选，用于回调或日志）
}
