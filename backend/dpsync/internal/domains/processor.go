package domains

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bitleak/lmstfy/client"
	"github.com/google/uuid"

	"oip/dpsync/internal/domains/common/job"
	"oip/dpsync/pkg/lmstfyx"
	"oip/dpsync/pkg/logger"
)

// GetProcess 返回核心处理函数（注入到 Processor）
// 注意：diagnosisService 可选参数，用于业务处理（如果为 nil，handler 需要自行处理）
func GetProcess(log logger.Logger, diagnosisService interface{}) lmstfyx.Proc {
	return func(ctx context.Context, lmstfyJob *client.Job) *lmstfyx.JobResp {
		startTime := time.Now()

		// 1. 解析 Job
		standardJob, meta, bizPayload, err := parseJob(ctx, lmstfyJob, log)
		if err != nil {
			log.Errorf(ctx, "[GetProcess] parseJob failed: %v", err)
			return &lmstfyx.JobResp{
				Action: lmstfyx.JobRespStatusBury,
				Data:   nil,
			}
		}

		// 2. 注入 TraceID 和依赖到 Context
		ctx = context.WithValue(ctx, "trace_id", meta.RequestID)
		ctx = context.WithValue(ctx, "action_type", standardJob.Payload.Data.ActionType)
		ctx = context.WithValue(ctx, "start_time", startTime)
		if diagnosisService != nil {
			ctx = context.WithValue(ctx, "diagnosis_service", diagnosisService)
		}

		log.Infof(ctx, "[GetProcess] Processing job: action_type=%s, request_id=%s, id=%s",
			meta.ActionType, meta.RequestID, meta.ID)

		// 3. 从 HandlerMap 获取 Handler
		handlerFunc, ok := HandlerMap[standardJob.Payload.Data.ActionType]
		if !ok {
			log.Errorf(ctx, "[GetProcess] handler not found for action_type: %s", standardJob.Payload.Data.ActionType)
			return &lmstfyx.JobResp{
				Action: lmstfyx.JobRespStatusBury,
				Data:   nil,
			}
		}

		// 4. 调用 Handler（捕获 panic）
		var resp *lmstfyx.JobResp
		func() {
			defer func() {
				if r := recover(); r != nil {
					log.Errorf(ctx, "[GetProcess] handler panic: %v", r)
					resp = &lmstfyx.JobResp{
						Action: lmstfyx.JobRespStatusBury,
						Data:   nil,
					}
				}
			}()

			handler, err := handlerFunc(ctx, meta, bizPayload)
			if err != nil {
				log.Errorf(ctx, "[GetProcess] handler creation failed: %v", err)
				resp = &lmstfyx.JobResp{
					Action: lmstfyx.JobRespStatusBury,
					Data:   nil,
				}
				return
			}

			handlerResp := handler.GetProcess()
			resp = doJobReport(ctx, handlerResp, meta, standardJob, startTime, lmstfyJob.ID, log)
		}()

		// 5. 记录处理时长
		duration := time.Since(startTime)
		log.Infof(ctx, "[GetProcess] Processing complete: action=%d, duration=%v", resp.Action, duration)

		return resp
	}
}

// parseJob 解析 Job
func parseJob(ctx context.Context, lmstfyJob *client.Job, log logger.Logger) (*job.Job, *job.Meta, interface{}, error) {
	// 1. 反序列化 Job
	var standardJob job.Job
	if err := json.Unmarshal(lmstfyJob.Data, &standardJob); err != nil {
		return nil, nil, nil, fmt.Errorf("json unmarshal failed: %w", err)
	}

	// 2. 校验必填字段
	if standardJob.Payload == nil || standardJob.Payload.Data == nil {
		return nil, nil, nil, fmt.Errorf("invalid job structure: payload.data is nil")
	}

	data := standardJob.Payload.Data

	// 3. 提取元数据
	meta := &job.Meta{
		RequestID:  data.RequestID,
		OrgID:      data.OrgID,
		ActionType: data.ActionType,
		ID:         data.ID,
	}

	// RequestID 为空则生成一个
	if meta.RequestID == "" {
		meta.RequestID = uuid.New().String()
	}

	// 4. 提取业务数据
	bizPayload := data.Data

	log.Debugf(ctx, "[parseJob] Parsed: action_type=%s, request_id=%s, id=%s",
		meta.ActionType, meta.RequestID, meta.ID)

	return &standardJob, meta, bizPayload, nil
}

// doJobReport 生成 JobResp（根据 Response 判断 ACK/Bury/Release）
func doJobReport(
	ctx context.Context,
	resp interface{},
	meta *job.Meta,
	standardJob *job.Job,
	startTime time.Time,
	jobID string,
	log logger.Logger,
) *lmstfyx.JobResp {
	// 序列化响应数据
	data, err := json.Marshal(resp)
	if err != nil {
		log.Errorf(ctx, "[doJobReport] marshal response failed: %v", err)
		return &lmstfyx.JobResp{
			Action: lmstfyx.JobRespStatusBury,
			Data:   nil,
		}
	}

	// TODO: 根据 resp.Error.Retryable 判断 Action（Phase 3 实现）
	// 目前默认返回 Success
	return &lmstfyx.JobResp{
		Action: lmstfyx.JobRespStatusSuccess,
		Data:   data,
	}
}
