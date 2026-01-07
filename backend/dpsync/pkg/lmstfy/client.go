package lmstfy

import (
	"fmt"
	"time"

	"github.com/bitleak/lmstfy/client"

	"oip/dpsync/internal/framework"
)

// Client Lmstfy 客户端封装
type Client struct {
	cli       *client.LmstfyClient
	namespace string
}

// NewClient 创建 Lmstfy 客户端
func NewClient(host string, port int, namespace string, token string) (*Client, error) {
	cli := client.NewLmstfyClient(host, port, namespace, token)
	return &Client{
		cli:       cli,
		namespace: namespace,
	}, nil
}

// Consume 消费消息（实现 MessageSource 接口）
func (c *Client) Consume(queue string, timeout time.Duration, ttr time.Duration) (*framework.Message, error) {
	// 将 timeout 转换为秒
	timeoutSec := uint32(timeout.Seconds())
	ttrSec := uint32(ttr.Seconds())

	// 调用 lmstfy 客户端
	job, err := c.cli.Consume(queue, timeoutSec, ttrSec)
	if err != nil {
		return nil, fmt.Errorf("lmstfy consume failed: %w", err)
	}

	// 超时未拉到消息
	if job == nil {
		return nil, nil
	}

	// 转换为框架 Message
	msg := &framework.Message{
		ID:    job.ID,
		Queue: job.Queue,
		Data:  job.Data,
		Extra: make(map[string]interface{}),
	}

	return msg, nil
}

// Ack 确认消息（实现 MessageSource 接口）
func (c *Client) Ack(queue string, jobID string) error {
	err := c.cli.Ack(queue, jobID)
	if err != nil {
		return fmt.Errorf("lmstfy ack failed: %w", err)
	}
	return nil
}

// Publish 发布消息
func (c *Client) Publish(queue string, data []byte, ttl, delay uint32) error {
	_, err := c.cli.Publish(queue, data, ttl, 3, delay)
	if err != nil {
		return fmt.Errorf("lmstfy publish failed: %w", err)
	}
	return nil
}
