package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"oip/common/model"
	"oip/dpmain/internal/app/domains/services/svcallback"
	"oip/dpmain/internal/app/infra/mq/lmstfy"
	"oip/dpmain/internal/app/pkg/logger"
)

// CallbackConsumer 回调消费者
// 职责：
// 1. 从 lmstfy 队列消费回调消息
// 2. 解析消息并调用 CallbackService 处理
// 3. 确认消息（ACK）
type CallbackConsumer struct {
	lmstfyClient    *lmstfy.Client
	callbackService *svcallback.CallbackService
	queueName       string
	logger          logger.Logger

	// 消费配置
	timeout      int // 拉取消息超时（秒）
	ttr          int // Time-To-Run（秒）
	pollInterval time.Duration
}

// Config 消费者配置
type Config struct {
	QueueName    string        // 队列名称
	Timeout      int           // 拉取消息超时（秒）
	TTR          int           // Time-To-Run（秒）
	PollInterval time.Duration // 轮询间隔
}

// NewCallbackConsumer 创建回调消费者实例
func NewCallbackConsumer(
	lmstfyClient *lmstfy.Client,
	callbackService *svcallback.CallbackService,
	config *Config,
	logger logger.Logger,
) *CallbackConsumer {
	return &CallbackConsumer{
		lmstfyClient:    lmstfyClient,
		callbackService: callbackService,
		queueName:       config.QueueName,
		timeout:         config.Timeout,
		ttr:             config.TTR,
		pollInterval:    config.PollInterval,
		logger:          logger,
	}
}

// Start 启动消费循环
func (c *CallbackConsumer) Start(ctx context.Context) error {
	c.logger.Info("Callback consumer started",
		"queue", c.queueName,
		"timeout", c.timeout,
		"ttr", c.ttr,
	)

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Callback consumer stopped")
			return ctx.Err()
		default:
			if err := c.consumeOne(ctx); err != nil {
				c.logger.Error("Failed to consume message", "error", err)
				time.Sleep(c.pollInterval)
			}
		}
	}
}

// consumeOne 消费一条消息
func (c *CallbackConsumer) consumeOne(ctx context.Context) error {
	// 1. 从队列拉取消息
	msg, err := c.lmstfyClient.Consume(ctx, c.queueName, c.timeout, c.ttr)
	if err != nil {
		return fmt.Errorf("consume message failed: %w", err)
	}

	if msg == nil {
		// 没有消息，继续等待
		return nil
	}

	c.logger.Info("Received callback message", "job_id", msg.JobID)

	// 2. 解析回调消息
	callback, err := c.parseMessage(msg.Data)
	if err != nil {
		c.logger.Error("Failed to parse message", "job_id", msg.JobID, "error", err)
		// 解析失败，直接 ACK（避免死循环）
		_ = c.lmstfyClient.Ack(ctx, c.queueName, msg.JobID)
		return err
	}

	// 3. 处理回调
	if err := c.callbackService.HandleCallback(ctx, callback); err != nil {
		c.logger.Error("Failed to handle callback",
			"job_id", msg.JobID,
			"order_id", callback.OrderID,
			"error", err,
		)
		// 处理失败，不 ACK（让 lmstfy TTR 机制重试）
		return err
	}

	// 4. 确认消息
	if err := c.lmstfyClient.Ack(ctx, c.queueName, msg.JobID); err != nil {
		c.logger.Error("Failed to ack message", "job_id", msg.JobID, "error", err)
		return err
	}

	c.logger.Info("Callback message processed successfully",
		"job_id", msg.JobID,
		"order_id", callback.OrderID,
	)

	return nil
}

// parseMessage 解析消息数据
// lmstfy 返回的消息可能是 Base64 编码的 URL form data
// 这里需要根据实际情况处理
func (c *CallbackConsumer) parseMessage(data json.RawMessage) (*model.OrderDiagnoseCallback, error) {
	var callback model.OrderDiagnoseCallback
	if err := json.Unmarshal(data, &callback); err != nil {
		return nil, fmt.Errorf("unmarshal callback failed: %w", err)
	}

	// 校验必填字段
	if callback.OrderID == "" {
		return nil, fmt.Errorf("order_id is required")
	}
	if callback.Status == "" {
		return nil, fmt.Errorf("status is required")
	}

	return &callback, nil
}
