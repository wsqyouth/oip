package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// PubSubClient Redis Pub/Sub 客户端封装
type PubSubClient struct {
	rdb *redis.Client
}

// NewPubSubClient 创建 Pub/Sub 客户端，支持密码认证
func NewPubSubClient(addr, password string, db int) (*PubSubClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return &PubSubClient{rdb: rdb}, nil
}

// Subscribe 订阅指定 channel 并等待消息，支持超时控制
// 用于 Smart Wait：订阅诊断结果频道，等待 dpsync 推送结果
func (c *PubSubClient) Subscribe(ctx context.Context, channel string, timeout time.Duration) (string, error) {
	sub := c.rdb.Subscribe(ctx, channel)
	defer sub.Close()

	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	select {
	case msg := <-sub.Channel():
		return msg.Payload, nil
	case <-timeoutCtx.Done():
		return "", timeoutCtx.Err()
	}
}

// Publish 向指定 channel 发布消息
func (c *PubSubClient) Publish(ctx context.Context, channel string, message string) error {
	return c.rdb.Publish(ctx, channel, message).Err()
}

// Close 关闭连接
func (c *PubSubClient) Close() error {
	return c.rdb.Close()
}
