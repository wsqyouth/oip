package worker

import (
	"context"
	"fmt"
	"sync"

	"go.uber.org/atomic"

	"oip/dpsync/internal/domains"
	"oip/dpsync/internal/framework"
	"oip/dpsync/pkg/config"
	"oip/dpsync/pkg/lmstfy"
	"oip/dpsync/pkg/logger"
)

// Manager 接口
type Manager interface {
	Start() error
	Shutdown()
}

// ManagerInstance Manager 实例
type ManagerInstance struct {
	ctx           context.Context
	cfg           *config.Config
	lmstfyClient  *lmstfy.Client
	callbackQueue string
	workers       []Worker
	closing       *atomic.Bool
	shutdownCh    chan struct{}
	wg            sync.WaitGroup
	mu            sync.RWMutex
	logger        logger.Logger
}

// NewManagerInstance 创建 Manager
func NewManagerInstance(cfg *config.Config, log logger.Logger) (Manager, error) {
	ctx := context.Background()

	// 初始化 lmstfy 客户端
	lmstfyClient, err := lmstfy.NewClient(cfg.Lmstfy.Host, cfg.Lmstfy.Port, cfg.Lmstfy.Namespace, cfg.Lmstfy.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to create lmstfy client: %w", err)
	}

	var callbackQueue string
	if len(cfg.Workers) > 0 {
		callbackQueue = cfg.Workers[0].CallbackQueue
	}
	if callbackQueue == "" {
		return nil, fmt.Errorf("callback_queue is required in worker config")
	}

	log.Infof(ctx, "[Manager] Initialized with callback_queue: %s", callbackQueue)

	return &ManagerInstance{
		ctx:           ctx,
		cfg:           cfg,
		lmstfyClient:  lmstfyClient,
		callbackQueue: callbackQueue,
		closing:       atomic.NewBool(false),
		shutdownCh:    make(chan struct{}),
		workers:       make([]Worker, 0),
		logger:        log,
	}, nil
}

// Start 启动 Manager
func (m *ManagerInstance) Start() error {
	m.logger.Infof(m.ctx, "[Manager] Starting...")

	// 1. 加载所有 Worker
	if err := m.loadWorkers(); err != nil {
		return fmt.Errorf("failed to load workers: %w", err)
	}

	m.logger.Infof(m.ctx, "[Manager] All workers loaded, count: %d", len(m.workers))

	// 2. 启动所有 Worker（每个 Worker 在独立 goroutine）
	for _, worker := range m.workers {
		w := worker
		m.wg.Add(1)
		go func() {
			defer m.wg.Done()
			w.Start()
		}()
		m.logger.Infof(m.ctx, "[Manager] Worker started: %s", w.GetName())
	}

	m.logger.Infof(m.ctx, "[Manager] Start success")

	// 3. 阻塞等待退出信号
	<-m.shutdownCh

	return nil
}

// Shutdown 优雅退出
func (m *ManagerInstance) Shutdown() {
	m.logger.Infof(m.ctx, "[Manager] Began to close")

	// 原子操作，保证并发安全
	if m.closing.CAS(false, true) {
		// 1. 所有 Worker 安全退出
		for _, worker := range m.workers {
			m.logger.Infof(m.ctx, "[Manager] Shutting down worker: %s", worker.GetName())
			worker.Shutdown()
		}

		// 2. 等待所有 Worker 退出
		m.wg.Wait()

		// 3. 关闭信号通道
		close(m.shutdownCh)

		m.logger.Infof(m.ctx, "[Manager] Shutdown complete")
	}
}

// loadWorkers 加载所有 Worker
func (m *ManagerInstance) loadWorkers() error {
	// 遍历配置中的所有 Worker
	for _, workerCfg := range m.cfg.Workers {
		// 创建 Subscriber 配置
		subCfg := &framework.SubscriberConfig{
			QueueName:    workerCfg.QueueName,
			Concurrency:  workerCfg.Subscriber.Threads,
			Rate:         workerCfg.Subscriber.Rate,
			Timeout:      workerCfg.Subscriber.Timeout,
			TTR:          workerCfg.Subscriber.TTR,
			ErrorBackoff: workerCfg.Subscriber.ErrorBackoff,
		}

		// 创建 Processor 配置
		procCfg := &framework.ProcessorConfig{
			Concurrency: workerCfg.Processor.Threads,
			BufferSize:  workerCfg.Processor.BufferSize,
			Timeout:     workerCfg.Processor.Timeout,
		}

		// 获取 GetProcess 函数
		getProcess := domains.GetProcess(m.logger, m.lmstfyClient, m.callbackQueue)

		// 创建 Worker 实例
		worker, err := NewWorkerInstance(
			m.ctx,
			workerCfg.Name,
			subCfg,
			procCfg,
			m.lmstfyClient, // MessageSource
			getProcess,     // lmstfyx.Proc
			m.logger,
		)
		if err != nil {
			return fmt.Errorf("failed to create worker %s: %w", workerCfg.Name, err)
		}

		m.workers = append(m.workers, worker)
	}

	return nil
}
