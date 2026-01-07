package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// PubSub Redis 发布/订阅客户端
type PubSub struct {
	client *redis.Client
}

// NewPubSub 创建 PubSub 实例
func NewPubSub(addr, password string, db int) (*PubSub, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// 测试连接
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &PubSub{
		client: client,
	}, nil
}

// DiagnosisNotification 诊断完成通知消息
type DiagnosisNotification struct {
	OrderID   string `json:"order_id"`
	AccountID int64  `json:"account_id"`
	Status    string `json:"status"` // DIAGNOSED/FAILED
	Timestamp int64  `json:"timestamp"`
}

// PublishDiagnosisComplete 发布诊断完成通知
// 参数：
//   - ctx: 上下文
//   - channel: Redis 频道名称（建议：order_diagnosis_complete）
//   - notification: 通知消息
func (p *PubSub) PublishDiagnosisComplete(
	ctx context.Context,
	channel string,
	notification *DiagnosisNotification,
) error {
	// 序列化通知消息
	msgJSON, err := json.Marshal(notification)
	if err != nil {
		return fmt.Errorf("failed to marshal notification: %w", err)
	}

	// 发布到 Redis 频道
	if err := p.client.Publish(ctx, channel, msgJSON).Err(); err != nil {
		return fmt.Errorf("failed to publish notification: %w", err)
	}

	return nil
}

// Subscribe 订阅 Redis 频道（用于测试）
func (p *PubSub) Subscribe(ctx context.Context, channel string) *redis.PubSub {
	return p.client.Subscribe(ctx, channel)
}

// Close 关闭 Redis 连接
func (p *PubSub) Close() error {
	return p.client.Close()
}
