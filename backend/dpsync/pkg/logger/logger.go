package logger

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 日志接口
type Logger interface {
	Debugf(ctx context.Context, format string, args ...interface{})
	Infof(ctx context.Context, format string, args ...interface{})
	Warnf(ctx context.Context, format string, args ...interface{})
	Errorf(ctx context.Context, format string, args ...interface{})
	Sync() error
}

// ZapLogger Zap 日志实现
type ZapLogger struct {
	logger *zap.Logger
}

// NewZapLogger 创建 Zap 日志实例
func NewZapLogger(level string) (Logger, error) {
	// 解析日志级别
	var zapLevel zapcore.Level
	switch level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// 配置
	cfg := zap.Config{
		Level:            zap.NewAtomicLevelAt(zapLevel),
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return &ZapLogger{logger: logger}, nil
}

// extractFields 从 Context 提取日志字段
func (l *ZapLogger) extractFields(ctx context.Context) []zap.Field {
	fields := make([]zap.Field, 0)

	// 提取 trace_id
	if traceID, ok := ctx.Value("trace_id").(string); ok && traceID != "" {
		fields = append(fields, zap.String("trace_id", traceID))
	}

	// 提取 worker_id
	if workerID, ok := ctx.Value("worker_id").(int); ok {
		fields = append(fields, zap.Int("worker_id", workerID))
	}

	// 提取 action_type
	if actionType, ok := ctx.Value("action_type").(string); ok && actionType != "" {
		fields = append(fields, zap.String("action_type", actionType))
	}

	return fields
}

// Debugf 输出 Debug 日志
func (l *ZapLogger) Debugf(ctx context.Context, format string, args ...interface{}) {
	fields := l.extractFields(ctx)
	l.logger.Debug(fmt.Sprintf(format, args...), fields...)
}

// Infof 输出 Info 日志
func (l *ZapLogger) Infof(ctx context.Context, format string, args ...interface{}) {
	fields := l.extractFields(ctx)
	l.logger.Info(fmt.Sprintf(format, args...), fields...)
}

// Warnf 输出 Warn 日志
func (l *ZapLogger) Warnf(ctx context.Context, format string, args ...interface{}) {
	fields := l.extractFields(ctx)
	l.logger.Warn(fmt.Sprintf(format, args...), fields...)
}

// Errorf 输出 Error 日志
func (l *ZapLogger) Errorf(ctx context.Context, format string, args ...interface{}) {
	fields := l.extractFields(ctx)
	l.logger.Error(fmt.Sprintf(format, args...), fields...)
}

// Sync 同步日志缓冲区
func (l *ZapLogger) Sync() error {
	return l.logger.Sync()
}
