package config

import (
	"fmt"

	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	App    AppConfig    `mapstructure:"app"`
	Server ServerConfig `mapstructure:"server"`
	MySQL  MySQLConfig  `mapstructure:"mysql"`
	Redis  RedisConfig  `mapstructure:"redis"`
	Lmstfy LmstfyConfig `mapstructure:"lmstfy"`
}

type AppConfig struct {
	Name     string `mapstructure:"name"`
	Env      string `mapstructure:"env"`
	LogLevel string `mapstructure:"log_level"`
}

type ServerConfig struct {
	Port string `mapstructure:"port"`
}

type MySQLConfig struct {
	DSN string `mapstructure:"dsn"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type LmstfyConfig struct {
	Host          string `mapstructure:"host"`
	Namespace     string `mapstructure:"namespace"`
	Queue         string `mapstructure:"queue"`
	CallbackQueue string `mapstructure:"callback_queue"`
	Token         string `mapstructure:"token"`
}

// Load 从配置文件加载配置
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

	// 兼容性处理：如果 server.port 为空，使用默认值
	if cfg.Server.Port == "" {
		cfg.Server.Port = "8080"
	}

	return &cfg, nil
}

// LoadDefault 加载默认配置文件路径
func LoadDefault() (*Config, error) {
	return Load("config/config.yaml")
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
	if c.Lmstfy.Token == "" {
		return fmt.Errorf("lmstfy token is required")
	}
	return nil
}

// GetServerPort 获取服务端口（兼容旧代码）
func (c *Config) GetServerPort() string {
	if c.Server.Port != "" {
		return c.Server.Port
	}
	return "8080"
}
