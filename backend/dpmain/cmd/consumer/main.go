package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"log"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"oip/dpmain/internal/app/config"
	"oip/dpmain/internal/app/infra/mq/lmstfy"
	"oip/dpmain/internal/app/infra/persistence/redis"
)

func main() {
	log.Println("Starting diagnosis consumer...")

	cfg := config.Load()

	// 初始化 lmstfy 客户端
	lmstfyClient := lmstfy.NewClient(cfg.Lmstfy.Host, cfg.Lmstfy.Namespace, "01KDCBF5BG0THBC24F1V53XPR1")

	// 初始化 Redis 客户端
	redisClient, err := redis.NewPubSubClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	log.Printf("Connected to lmstfy: %s", cfg.Lmstfy.Host)
	log.Printf("Connected to Redis: %s", cfg.Redis.Addr)
	log.Printf("Consuming from queue: %s", cfg.Lmstfy.Queue)

	// 启动消费循环
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 处理优雅退出
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received shutdown signal, stopping consumer...")
		cancel()
	}()

	// 消费循环
	for {
		select {
		case <-ctx.Done():
			log.Println("Consumer stopped")
			return
		default:
			if err := consumeOne(ctx, lmstfyClient, redisClient, cfg.Lmstfy.Queue); err != nil {
				log.Printf("[ERROR] Failed to consume message: %v", err)
				time.Sleep(1 * time.Second)
			}
		}
	}
}

// consumeOne 消费一条消息并处理
func consumeOne(ctx context.Context, lmstfyClient *lmstfy.Client, redisClient *redis.PubSubClient, queue string) error {
	// 从队列中获取消息（超时3秒，处理时间30秒）
	msg, err := lmstfyClient.Consume(ctx, queue, 3, 30)
	if err != nil {
		return err
	}

	if msg == nil {
		// 没有消息，继续等待
		return nil
	}

	log.Printf("[INFO] Received message: job_id=%s", msg.JobID)

	// 解析消息数据
	// lmstfy 返回的 data 字段是 Base64 编码的 URL form data
	var dataStr string
	if err := json.Unmarshal(msg.Data, &dataStr); err != nil {
		log.Printf("[ERROR] Failed to parse data string: %v", err)
		return lmstfyClient.Ack(ctx, queue, msg.JobID)
	}

	// Base64 解码
	decodedData, err := base64.StdEncoding.DecodeString(dataStr)
	if err != nil {
		log.Printf("[ERROR] Failed to decode base64: %v", err)
		return lmstfyClient.Ack(ctx, queue, msg.JobID)
	}

	// 解析 URL form data，提取 data 字段
	formData, err := url.ParseQuery(string(decodedData))
	if err != nil {
		log.Printf("[ERROR] Failed to parse form data: %v", err)
		return lmstfyClient.Ack(ctx, queue, msg.JobID)
	}

	jsonData := formData.Get("data")
	if jsonData == "" {
		log.Printf("[ERROR] Missing data field in form data")
		return lmstfyClient.Ack(ctx, queue, msg.JobID)
	}

	// URL decode JSON 数据
	jsonData, err = url.QueryUnescape(jsonData)
	if err != nil {
		log.Printf("[ERROR] Failed to unescape JSON data: %v", err)
		return lmstfyClient.Ack(ctx, queue, msg.JobID)
	}

	var orderData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonData), &orderData); err != nil {
		log.Printf("[ERROR] Failed to parse JSON data: %v, json=%s", err, jsonData)
		return lmstfyClient.Ack(ctx, queue, msg.JobID)
	}

	orderID, ok := orderData["order_id"].(string)
	if !ok {
		log.Printf("[ERROR] Missing order_id in message")
		return lmstfyClient.Ack(ctx, queue, msg.JobID)
	}

	log.Printf("[INFO] Processing order: %s", orderID)

	// 模拟诊断处理（2秒）
	time.Sleep(2 * time.Second)

	// 生成模拟的诊断结果
	diagnosisResult := generateMockDiagnosisResult()

	// 将诊断结果推送到 Redis
	channel := "diagnosis:result:" + orderID
	resultJSON, err := json.Marshal(diagnosisResult)
	if err != nil {
		log.Printf("[ERROR] Failed to marshal diagnosis result: %v", err)
		return lmstfyClient.Ack(ctx, queue, msg.JobID)
	}

	if err := redisClient.Publish(ctx, channel, string(resultJSON)); err != nil {
		log.Printf("[ERROR] Failed to publish diagnosis result to Redis: %v", err)
		return lmstfyClient.Ack(ctx, queue, msg.JobID)
	}

	log.Printf("[SUCCESS] Published diagnosis result for order %s to channel %s", orderID, channel)

	// 确认消息已处理
	if err := lmstfyClient.Ack(ctx, queue, msg.JobID); err != nil {
		log.Printf("[ERROR] Failed to ack message: %v", err)
		return err
	}

	log.Printf("[INFO] Message acknowledged: job_id=%s", msg.JobID)

	return nil
}

// generateMockDiagnosisResult 生成模拟的诊断结果
func generateMockDiagnosisResult() map[string]interface{} {
	return map[string]interface{}{
		"items": []map[string]interface{}{
			{
				"type":   "shipping",
				"status": "SUCCESS",
				"data_json": map[string]interface{}{
					"recommended_code": "FEDEX_GROUND",
					"rates": []map[string]interface{}{
						{
							"carrier":  "FedEx",
							"service":  "Ground",
							"price":    12.50,
							"currency": "USD",
						},
						{
							"carrier":  "USPS",
							"service":  "Priority Mail",
							"price":    10.00,
							"currency": "USD",
						},
					},
				},
			},
			{
				"type":   "anomaly",
				"status": "SUCCESS",
				"data_json": map[string]interface{}{
					"has_risk": false,
					"issues":   []string{},
				},
			},
		},
	}
}
