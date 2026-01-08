package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"oip/dpmain/internal/app/config"
	"oip/dpmain/internal/app/consumer"
	"oip/dpmain/internal/app/domains/repo/rporder"
	"oip/dpmain/internal/app/domains/services/svcallback"
	"oip/dpmain/internal/app/infra/mq/lmstfy"

	"oip/dpmain/internal/app/infra/persistence/redis"
	"oip/dpmain/internal/app/pkg/logger"
)

func main() {
	// 1. 加载配置
	cfg, err := config.LoadDefault()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 2. 初始化日志
	appLogger := logger.NewDefaultLogger()
	appLogger.Info("Starting callback consumer...")

	// 3. 初始化基础设施组件
	// 初始化数据库
	db, err := gorm.Open(mysql.Open(cfg.MySQL.DSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get sql.DB: %v", err)
	}
	defer sqlDB.Close()
	appLogger.Info("Database connected")

	// 初始化 Redis
	redisClient, err := redis.NewPubSubClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB)
	if err != nil {
		log.Fatalf("Failed to init redis: %v", err)
	}
	defer redisClient.Close()
	appLogger.Info("Redis connected")

	// 初始化 Lmstfy
	lmstfyClient := lmstfy.NewClient(cfg.Lmstfy.Host, cfg.Lmstfy.Namespace, cfg.Lmstfy.Token)
	appLogger.Info("Lmstfy client initialized")

	// 4. 初始化 Repository 层
	orderRepo := rporder.NewOrderRepository(db)

	// 5. 初始化 Service 层
	callbackService := svcallback.NewCallbackService(
		orderRepo,
		redisClient,
		appLogger,
	)

	// 6. 初始化 Consumer
	callbackConsumer := consumer.NewCallbackConsumer(
		lmstfyClient,
		callbackService,
		&consumer.Config{
			QueueName:    cfg.Lmstfy.CallbackQueue,
			Timeout:      3,  // 拉取消息超时 3 秒
			TTR:          30, // 消息处理超时 30 秒
			PollInterval: 100 * time.Millisecond,
		},
		appLogger,
	)

	// 7. 启动消费循环（优雅退出）
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 启动消费者（goroutine）
	errChan := make(chan error, 1)
	go func() {
		errChan <- callbackConsumer.Start(ctx)
	}()

	// 等待退出信号或错误
	select {
	case <-sigChan:
		appLogger.Info("Received shutdown signal, stopping consumer...")
		cancel()
		time.Sleep(1 * time.Second) // 等待消费者优雅退出
		appLogger.Info("Consumer stopped gracefully")
	case err := <-errChan:
		if err != nil && err != context.Canceled {
			appLogger.Error("Consumer stopped with error", "error", err)
			os.Exit(1)
		}
	}
}
