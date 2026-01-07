package lmstfy

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Client Lmstfy 客户端封装
type Client struct {
	host      string
	namespace string
	token     string
}

// NewClient 创建 Lmstfy 客户端
func NewClient(host, namespace, token string) *Client {
	return &Client{
		host:      strings.TrimSuffix(host, "/"),
		namespace: namespace,
		token:     token,
	}
}

// Publish 发布消息到队列
// TTL: 消息存活时间（秒），Delay: 延迟时间（秒），Tries: 重试次数
func (c *Client) Publish(ctx context.Context, queue string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// 修复：使用 query 参数方式，直接将 JSON bytes 作为 body 发送
	// 这样与官方 lmstfy Go 客户端保持一致
	endpoint := fmt.Sprintf("%s/api/%s/%s?ttl=3600&delay=0&tries=3", c.host, c.namespace, queue)

	req, err := http.NewRequestWithContext(ctx, "PUT", endpoint, strings.NewReader(string(payload)))
	if err != nil {
		return err
	}

	// 不设置 Content-Type，让 lmstfy 接受原始 body
	if c.token != "" {
		req.Header.Set("X-Token", c.token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("lmstfy publish failed: status=%d", resp.StatusCode)
	}

	return nil
}

// Message 队列消息结构
type Message struct {
	JobID string          `json:"job_id"`
	Data  json.RawMessage `json:"data"`
}

// Consume 从队列中消费消息
// timeout: 等待超时时间（秒），ttr: 消息处理超时时间（秒）
func (c *Client) Consume(ctx context.Context, queue string, timeout, ttr int) (*Message, error) {
	endpoint := fmt.Sprintf("%s/api/%s/%s?timeout=%d&ttr=%d", c.host, c.namespace, queue, timeout, ttr)

	req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	if c.token != "" {
		req.Header.Set("X-Token", c.token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// 队列为空，没有消息
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("lmstfy consume failed: status=%d", resp.StatusCode)
	}

	// lmstfy HTTP API 返回的消息格式 (response from lmstfy includes base64-encoded data)
	type LmstfyResponse struct {
		JobID string `json:"job_id"`
		Data  string `json:"data"` // base64 encoded
	}

	var lmstfyResp LmstfyResponse
	if err := json.NewDecoder(resp.Body).Decode(&lmstfyResp); err != nil {
		return nil, err
	}

	// Base64 decode the data field
	decodedData, err := base64.StdEncoding.DecodeString(lmstfyResp.Data)
	if err != nil {
		return nil, fmt.Errorf("failed to base64 decode message data: %w", err)
	}

	msg := &Message{
		JobID: lmstfyResp.JobID,
		Data:  json.RawMessage(decodedData),
	}

	return msg, nil
}

// Ack 确认消息已处理
func (c *Client) Ack(ctx context.Context, queue, jobID string) error {
	endpoint := fmt.Sprintf("%s/api/%s/%s/job/%s", c.host, c.namespace, queue, jobID)

	req, err := http.NewRequestWithContext(ctx, "DELETE", endpoint, nil)
	if err != nil {
		return err
	}

	if c.token != "" {
		req.Header.Set("X-Token", c.token)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("lmstfy ack failed: status=%d", resp.StatusCode)
	}

	return nil
}
