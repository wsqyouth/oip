package framework

import (
	"context"
	"time"
)

// MessageSource 消息源接口（适配不同 MQ）
type MessageSource interface {
	// Consume 消费消息（阻塞，直到拉取到消息或超时）
	Consume(queue string, timeout time.Duration, ttr time.Duration) (*Message, error)

	// Ack 确认消息（删除消息）
	Ack(queue string, jobID string) error
}

// Logger 日志接口
type Logger interface {
	Debugf(ctx context.Context, format string, args ...interface{})
	Infof(ctx context.Context, format string, args ...interface{})
	Warnf(ctx context.Context, format string, args ...interface{})
	Errorf(ctx context.Context, format string, args ...interface{})
}
