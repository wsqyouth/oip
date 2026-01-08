package main

// @title           OIP Backend API
// @version         1.0
// @description     跨境订单智能诊断平台后端 API，提供订单接入和智能诊断服务
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.email  support@oip.example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name api-key
// @description API Key 用于接口认证（当前版本暂未启用，保留占位）

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"oip/dpmain/internal/app/config"
)

func main() {
	// 1. 加载配置
	cfg, err := config.LoadDefault()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Config validation failed: %v", err)
	}

	// 2. 初始化应用（包含 HTTP Server 和 Consumer）
	app, cleanup, err := InitializeApp(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}
	defer cleanup()

	// 3. 创建 HTTP Server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: app.Engine,
	}

	// 4. 启动 Consumer（后台 goroutine）
	consumerCtx, cancelConsumer := context.WithCancel(context.Background())
	consumerErrChan := make(chan error, 1)

	go func() {
		log.Printf("Starting callback consumer...")
		consumerErrChan <- app.CallbackConsumer.Start(consumerCtx)
	}()

	// 5. 启动 HTTP Server（后台 goroutine）
	serverErrChan := make(chan error, 1)
	go func() {
		log.Printf("Starting HTTP server on %s", addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrChan <- err
		}
	}()

	// 6. 优雅停机处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-sigChan:
		log.Println("Received shutdown signal, gracefully shutting down...")
		gracefulShutdown(server, cancelConsumer)
	case err := <-serverErrChan:
		log.Fatalf("HTTP server error: %v", err)
	case err := <-consumerErrChan:
		if err != nil && err != context.Canceled {
			log.Fatalf("Consumer error: %v", err)
		}
	}

	log.Println("Application stopped")
}

// gracefulShutdown 优雅停机
func gracefulShutdown(server *http.Server, cancelConsumer context.CancelFunc) {
	// 1. 停止 Consumer
	log.Println("Stopping consumer...")
	cancelConsumer()
	time.Sleep(1 * time.Second) // 等待消费者处理完当前消息

	// 2. 停止 HTTP Server
	log.Println("Stopping HTTP server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	} else {
		log.Println("HTTP server stopped gracefully")
	}

	log.Println("All services stopped gracefully")
}
