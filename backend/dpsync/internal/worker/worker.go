package worker

import (
	"context"

	"oip/dpsync/internal/framework"
	"oip/dpsync/pkg/lmstfyx"
	"oip/dpsync/pkg/logger"
)

// Worker 接口
type Worker interface {
	Start()
	Shutdown()
	GetName() string
}

// WorkerInstance Worker 实例
type WorkerInstance struct {
	ctx        context.Context
	name       string
	subscriber *framework.Subscriber
	processor  *framework.Processor
	inputChan  chan *framework.Message
	shutdownCh chan struct{}
	logger     logger.Logger
}

// NewWorkerInstance 创建 Worker 实例
func NewWorkerInstance(
	ctx context.Context,
	name string,
	subscriberCfg *framework.SubscriberConfig,
	processorCfg *framework.ProcessorConfig,
	source framework.MessageSource,
	proc lmstfyx.Proc, // 注入 GetProcess
	log logger.Logger,
) (Worker, error) {
	// 创建 inputChan（缓冲区）
	inputChan := make(chan *framework.Message, processorCfg.BufferSize)

	// 创建 Subscriber
	subscriber := framework.NewSubscriber(subscriberCfg, source, log)

	// 创建 Processor
	processor := framework.NewProcessor(processorCfg, proc, log)

	return &WorkerInstance{
		ctx:        ctx,
		name:       name,
		subscriber: subscriber,
		processor:  processor,
		inputChan:  inputChan,
		shutdownCh: make(chan struct{}),
		logger:     log,
	}, nil
}

// Start 启动 Worker
func (w *WorkerInstance) Start() {
	w.logger.Infof(w.ctx, "[Worker] %s started", w.name)

	// 1. 启动 Processor
	w.processor.Start(w.ctx, w.inputChan)

	// 2. 启动 Subscriber
	w.subscriber.Start(w.ctx, w.inputChan)

	// 3. 阻塞，等待关闭指令
	<-w.shutdownCh
}

// Shutdown 优雅退出（4 步链路）
func (w *WorkerInstance) Shutdown() {
	w.logger.Infof(w.ctx, "[Worker] %s began to close", w.name)

	// 【第 1 步】停止拉取新消息
	w.subscriber.Stop()

	// 【第 2 步】等待 Subscriber 完全退出
	w.subscriber.Wait()

	// 【第 3 步】通知 Processor 进入 Drain 模式
	w.processor.SignalShutdown()

	// 【第 4 步】等待 Processor 处理完剩余消息
	w.processor.Wait()

	close(w.shutdownCh)
	w.logger.Infof(w.ctx, "[Worker] %s shutdown complete", w.name)
}

// GetName 获取 Worker 名称
func (w *WorkerInstance) GetName() string {
	return w.name
}
