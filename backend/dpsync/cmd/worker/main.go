package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"oip/dpsync/internal/worker"
	"oip/dpsync/pkg/config"
	"oip/dpsync/pkg/logger"
)

var (
	configPath = flag.String("config", "./config/worker.yaml", "配置文件路径")
)

func main() {
	flag.Parse()

	// 1. 初始化日志
	log.Println("========================================")
	log.Println("  DPSYNC Worker Starting...")
	log.Println("========================================")

	// 2. 加载配置
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Config validation failed: %v", err)
	}

	log.Printf("Config loaded: %s, env: %s, log_level: %s\n", cfg.App.Name, cfg.App.Env, cfg.App.LogLevel)

	// 3. 初始化 Logger
	zapLogger, err := logger.NewZapLogger(cfg.App.LogLevel)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer zapLogger.Sync()

	// 4. 创建 Manager
	mgr, err := worker.NewManagerInstance(cfg, zapLogger)
	if err != nil {
		log.Fatalf("Failed to create manager: %v", err)
	}

	// 5. 启动 Manager（goroutine）
	go func() {
		if err := mgr.Start(); err != nil {
			log.Fatalf("Manager start failed: %v", err)
		}
	}()

	log.Println("Worker started. Press Ctrl+C to shutdown.")

	// 6. 等待退出信号
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh

	log.Println("========================================")
	log.Printf("  Received signal: %v\n", sig)
	log.Println("  Shutting down Worker...")
	log.Println("========================================")

	// 7. 优雅关闭 Manager
	mgr.Shutdown()

	fmt.Println("========================================")
	fmt.Println("  Worker exited gracefully")
	fmt.Println("========================================")
}
