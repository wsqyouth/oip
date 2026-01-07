package logger

import (
	"context"
	"log"
	"os"
)

// Logger 日志接口（预留，可扩展为 zap/logrus）
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Debug(msg string, fields ...interface{})

	// Context 支持（用于链路追踪）
	InfoContext(ctx context.Context, msg string, fields ...interface{})
	ErrorContext(ctx context.Context, msg string, fields ...interface{})
	WarnContext(ctx context.Context, msg string, fields ...interface{})
	DebugContext(ctx context.Context, msg string, fields ...interface{})
}

// DefaultLogger 默认日志实现
type DefaultLogger struct {
	logger *log.Logger
}

// NewDefaultLogger 创建默认日志
func NewDefaultLogger() *DefaultLogger {
	return &DefaultLogger{
		logger: log.New(os.Stdout, "[DPMAIN] ", log.LstdFlags),
	}
}

func (l *DefaultLogger) Info(msg string, fields ...interface{}) {
	l.logger.Printf("[INFO] "+msg, fields...)
}

func (l *DefaultLogger) Error(msg string, fields ...interface{}) {
	l.logger.Printf("[ERROR] "+msg, fields...)
}

func (l *DefaultLogger) Warn(msg string, fields ...interface{}) {
	l.logger.Printf("[WARN] "+msg, fields...)
}

func (l *DefaultLogger) Debug(msg string, fields ...interface{}) {
	l.logger.Printf("[DEBUG] "+msg, fields...)
}

func (l *DefaultLogger) InfoContext(ctx context.Context, msg string, fields ...interface{}) {
	l.logger.Printf("[INFO] "+msg, fields...)
}

func (l *DefaultLogger) ErrorContext(ctx context.Context, msg string, fields ...interface{}) {
	l.logger.Printf("[ERROR] "+msg, fields...)
}

func (l *DefaultLogger) WarnContext(ctx context.Context, msg string, fields ...interface{}) {
	l.logger.Printf("[WARN] "+msg, fields...)
}

func (l *DefaultLogger) DebugContext(ctx context.Context, msg string, fields ...interface{}) {
	l.logger.Printf("[DEBUG] "+msg, fields...)
}
