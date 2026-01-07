package config

import (
	"fmt"
	"os"
)

// Config 应用配置
type Config struct {
	Server ServerConfig
	MySQL  MySQLConfig
	Redis  RedisConfig
	Lmstfy LmstfyConfig
}

type ServerConfig struct {
	Port string
}

type MySQLConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type LmstfyConfig struct {
	Host          string
	Namespace     string
	Queue         string
	CallbackQueue string // 回调队列名称
	Token         string // Lmstfy Token
}

// Load 从环境变量加载配置
func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
		},
		MySQL: MySQLConfig{
			DSN: getEnv("MYSQL_DSN", "root:password@tcp(127.0.0.1:3306)/oip?parseTime=true&loc=Local"),
		},
		Redis: RedisConfig{
			Addr:     getEnv("REDIS_ADDR", "localhost:6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       0,
		},
		Lmstfy: LmstfyConfig{
			Host:          getEnv("LMSTFY_HOST", "http://localhost:7777"),
			Namespace:     getEnv("LMSTFY_NAMESPACE", "oip"),
			Queue:         getEnv("LMSTFY_QUEUE", "order_diagnose"),
			CallbackQueue: getEnv("LMSTFY_CALLBACK_QUEUE", "order_diagnose_callback"),
			Token:         getEnv("LMSTFY_TOKEN", "01KDCBF5BG0THBC24F1V53XPR1"),
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Validate 验证配置完整性
func (c *Config) Validate() error {
	if c.MySQL.DSN == "" {
		return fmt.Errorf("mysql dsn is required")
	}
	if c.Redis.Addr == "" {
		return fmt.Errorf("redis addr is required")
	}
	if c.Lmstfy.Host == "" {
		return fmt.Errorf("lmstfy host is required")
	}
	return nil
}
