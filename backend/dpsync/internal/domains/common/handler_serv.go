package common

import (
	"context"

	"oip/dpsync/internal/domains/common/job"
	"oip/dpsync/internal/domains/common/response"
)

// HandlerServProc Handler 构造函数类型
type HandlerServProc func(ctx context.Context, meta *job.Meta, payload interface{}) (HandlerServ, error)

// HandlerServ Handler 接口
type HandlerServ interface {
	GetProcess() *response.Response
}
