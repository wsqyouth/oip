package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 全局配置
type Config struct {
	App     AppConfig      `mapstructure:"app"`
	MySQL   MySQLConfig    `mapstructure:"mysql"`
	Redis   RedisConfig    `mapstructure:"redis"`
	Lmstfy  LmstfyConfig   `mapstructure:"lmstfy"`
	Workers []WorkerConfig `mapstructure:"workers"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name     string `mapstructure:"name"`
	Env      string `mapstructure:"env"`
	LogLevel string `mapstructure:"log_level"`
}

// MySQLConfig MySQL 配置
type MySQLConfig struct {
	DSN string `mapstructure:"dsn"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// LmstfyConfig Lmstfy 配置
type LmstfyConfig struct {
	Host      string `mapstructure:"host"`
	Port      int    `mapstructure:"port"`
	Namespace string `mapstructure:"namespace"`
	Token     string `mapstructure:"token"`
}

// WorkerConfig Worker 配置
type WorkerConfig struct {
	Name          string           `mapstructure:"name"`
	QueueName     string           `mapstructure:"queue_name"`
	CallbackQueue string           `mapstructure:"callback_queue"` // 回调队列名称
	Subscriber    SubscriberConfig `mapstructure:"subscriber"`
	Processor     ProcessorConfig  `mapstructure:"processor"`
}

// SubscriberConfig Subscriber 配置
type SubscriberConfig struct {
	Threads      int           `mapstructure:"threads"`       // 并发拉取数
	Rate         time.Duration `mapstructure:"rate"`          // 拉取速率
	Timeout      time.Duration `mapstructure:"timeout"`       // 拉取超时
	TTR          time.Duration `mapstructure:"ttr"`           // Time-To-Run
	ErrorBackoff time.Duration `mapstructure:"error_backoff"` // 错误退避时间
}

// ProcessorConfig Processor 配置
type ProcessorConfig struct {
	Threads    int           `mapstructure:"threads"`     // 并发处理数
	BufferSize int           `mapstructure:"buffer_size"` // Channel 缓冲大小
	Timeout    time.Duration `mapstructure:"timeout"`     // 单个任务超时
}

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config failed: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config failed: %w", err)
	}

	return &cfg, nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.App.Name == "" {
		return fmt.Errorf("app.name is required")
	}
	if c.Lmstfy.Host == "" {
		return fmt.Errorf("lmstfy.host is required")
	}
	if len(c.Workers) == 0 {
		return fmt.Errorf("at least one worker is required")
	}
	return nil
}
