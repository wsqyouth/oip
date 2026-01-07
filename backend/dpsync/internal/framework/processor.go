package framework

import (
	"context"
	"sync"
	"time"

	"github.com/bitleak/lmstfy/client"

	"oip/dpsync/pkg/lmstfyx"
)

// Processor 处理器：接收消息，调用业务处理函数
type Processor struct {
	cfg        *ProcessorConfig
	proc       lmstfyx.Proc // 业务处理函数（注入的 GetProcess）
	logger     Logger
	shutdownCh chan struct{} // 专门的退出信号通道
	wg         sync.WaitGroup
}

// NewProcessor 创建处理器
func NewProcessor(cfg *ProcessorConfig, proc lmstfyx.Proc, logger Logger) *Processor {
	return &Processor{
		cfg:        cfg,
		proc:       proc,
		logger:     logger,
		shutdownCh: make(chan struct{}),
	}
}

// Start 启动处理协程
func (p *Processor) Start(ctx context.Context, inputChan <-chan *Message) error {
	p.logger.Infof(ctx, "[Processor] Starting with %d workers", p.cfg.Concurrency)

	for i := 0; i < p.cfg.Concurrency; i++ {
		workerID := i
		p.wg.Add(1)
		go p.loop(ctx, workerID, inputChan)
	}

	return nil
}

// SignalShutdown 通知 Processor 准备退出（进入 Drain 模式）
func (p *Processor) SignalShutdown() {
	p.logger.Infof(context.Background(), "[Processor] Shutdown signal received")
	close(p.shutdownCh) // 关闭信号通道
}

// Wait 等待所有处理协程退出
func (p *Processor) Wait() {
	p.wg.Wait()
	p.logger.Infof(context.Background(), "[Processor] All workers exited")
}

// loop 处理循环（单个 Worker）
func (p *Processor) loop(ctx context.Context, workerID int, inputChan <-chan *Message) {
	defer p.wg.Done()
	p.logger.Infof(ctx, "[Processor-%d] Started", workerID)

	for {
		select {
		// A. 正常业务处理
		case msg := <-inputChan:
			p.process(ctx, msg, workerID)

		// B. Drain 模式：处理完剩余消息再退出
		case <-p.shutdownCh:
			p.logger.Infof(ctx, "[Processor-%d] Entering DRAIN mode", workerID)
			count := 0
			for {
				select {
				case msg := <-inputChan:
					p.process(ctx, msg, workerID)
					count++
				default:
					// Channel 空了，安全退出
					p.logger.Infof(ctx, "[Processor-%d] Drained %d messages, exiting", workerID, count)
					return
				}
			}
		}
	}
}

// process 处理单个消息
func (p *Processor) process(ctx context.Context, msg *Message, workerID int) {
	if msg == nil {
		return
	}

	startTime := time.Now()

	// 1. 创建超时控制的 Context
	procCtx, cancel := context.WithTimeout(ctx, p.cfg.Timeout)
	defer cancel()

	// 2. 注入元信息到 Context
	procCtx = context.WithValue(procCtx, "worker_id", workerID)
	procCtx = context.WithValue(procCtx, "message_id", msg.ID)
	procCtx = context.WithValue(procCtx, "start_time", startTime)

	p.logger.Infof(procCtx, "[Processor-%d] Processing message: %s", workerID, msg.ID)

	// 3. 调用业务处理函数（注入的 GetProcess）
	// 构造 lmstfy Job
	job := &client.Job{
		ID:    msg.ID,
		Queue: msg.Queue,
		Data:  msg.Data,
	}

	resp := p.proc(procCtx, job)

	// 4. 记录处理时长
	duration := time.Since(startTime)
	p.logger.Infof(procCtx, "[Processor-%d] Message processed: %s, action: %d, duration: %v",
		workerID, msg.ID, resp.Action, duration)

	// TODO: 根据 resp.Action 执行 ACK/Bury/Release（Phase 3 实现）
}
