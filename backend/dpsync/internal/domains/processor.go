package domains

import (
	"context"
	"time"

	"github.com/bitleak/lmstfy/client"

	"oip/dpsync/internal/framework"
	"oip/dpsync/pkg/lmstfy"
	"oip/dpsync/pkg/lmstfyx"
	"oip/dpsync/pkg/logger"
)

// GetProcess 返回核心处理函数
func GetProcess(log logger.Logger, lmstfyClient *lmstfy.Client, callbackQueue string) lmstfyx.Proc {
	return func(ctx context.Context, lmstfyJob *client.Job) *lmstfyx.JobResp {
		startTime := time.Now()

		baseHandler := &framework.BaseHandler{}
		if err := baseHandler.ParseJob(ctx, lmstfyJob.Data); err != nil {
			log.Errorf(ctx, "[GetProcess] parseJob failed: %v", err)
			return &lmstfyx.JobResp{
				Action: lmstfyx.JobRespStatusBury,
				Data:   nil,
			}
		}

		meta := baseHandler.GetMeta()
		log.Infof(ctx, "[GetProcess] Processing job: action_type=%s, request_id=%s, id=%s",
			meta.ActionType, meta.RequestID, meta.ID)

		handlerFactory, ok := HandlerMap[meta.ActionType]
		if !ok {
			log.Errorf(ctx, "[GetProcess] handler not found for action_type: %s", meta.ActionType)
			return &lmstfyx.JobResp{
				Action: lmstfyx.JobRespStatusBury,
				Data:   nil,
			}
		}

		handler, err := handlerFactory(ctx, baseHandler, lmstfyClient, callbackQueue)
		if err != nil {
			log.Errorf(ctx, "[GetProcess] handler creation failed: %v", err)
			return &lmstfyx.JobResp{
				Action: lmstfyx.JobRespStatusBury,
				Data:   nil,
			}
		}

		data, err := handler.Handle(ctx)
		if err != nil {
			log.Errorf(ctx, "[GetProcess] handler.Handle failed: %v", err)
			return &lmstfyx.JobResp{
				Action: lmstfyx.JobRespStatusBury,
				Data:   data,
			}
		}

		duration := time.Since(startTime)
		log.Infof(ctx, "[GetProcess] Processing complete: duration=%v", duration)

		return &lmstfyx.JobResp{
			Action: lmstfyx.JobRespStatusSuccess,
			Data:   data,
		}
	}
}
