package framework

import (
	"context"
	"sync"
	"time"
)

// Subscriber 订阅者：从消息队列拉取消息，转发给 Processor
type Subscriber struct {
	cfg        *SubscriberConfig
	source     MessageSource // 消息源（lmstfy 适配器）
	logger     Logger
	cancelFunc context.CancelFunc // 取消函数
	wg         sync.WaitGroup
}

// NewSubscriber 创建订阅者
func NewSubscriber(cfg *SubscriberConfig, source MessageSource, logger Logger) *Subscriber {
	return &Subscriber{
		cfg:    cfg,
		source: source,
		logger: logger,
	}
}

// Start 启动订阅循环
func (s *Subscriber) Start(parentCtx context.Context, inputChan chan<- *Message) error {
	// 核心：从父 Context 派生子 Context
	ctx, cancel := context.WithCancel(parentCtx)
	s.cancelFunc = cancel

	s.logger.Infof(ctx, "[Subscriber] Starting with %d workers for queue: %s",
		s.cfg.Concurrency, s.cfg.QueueName)

	// 启动多个并发拉取协程
	for i := 0; i < s.cfg.Concurrency; i++ {
		workerID := i
		s.wg.Add(1)
		go s.loop(ctx, workerID, inputChan)
	}

	return nil
}

// Stop 停止订阅（不再拉取新消息）
func (s *Subscriber) Stop() {
	s.logger.Infof(context.Background(), "[Subscriber] Stopping...")
	if s.cancelFunc != nil {
		s.cancelFunc() // 触发 ctx.Done()
	}
}

// Wait 等待所有订阅协程退出
func (s *Subscriber) Wait() {
	s.wg.Wait() // 关键：确保所有拉取协程退出
	s.logger.Infof(context.Background(), "[Subscriber] All workers exited")
}

// loop 订阅循环（单个 Worker）
func (s *Subscriber) loop(ctx context.Context, workerID int, inputChan chan<- *Message) {
	defer s.wg.Done()
	s.logger.Infof(ctx, "[Subscriber-%d] Started", workerID)

	for {
		// 1. 拉取消息（带超时）
		msg, err := s.source.Consume(s.cfg.QueueName, s.cfg.Timeout, s.cfg.TTR)
		if err != nil {
			// 容错：网络抖动不退出，只记录日志
			s.logger.Warnf(ctx, "[Subscriber-%d] Consume error: %v, retrying...", workerID, err)

			// 退出检查（即使出错也要检查是否该退出）
			select {
			case <-ctx.Done():
				s.logger.Infof(ctx, "[Subscriber-%d] Context cancelled, exiting", workerID)
				return
			default:
				// 错误退避
				time.Sleep(s.cfg.ErrorBackoff)
				continue
			}
		}

		// nil 消息（超时未拉到），继续循环
		if msg == nil {
			select {
			case <-ctx.Done():
				s.logger.Infof(ctx, "[Subscriber-%d] Context cancelled, exiting", workerID)
				return
			default:
				continue
			}
		}

		// 2. 发送给 Processor（防死锁设计）
		select {
		case inputChan <- msg:
			// 发送成功
			s.logger.Debugf(ctx, "[Subscriber-%d] Message sent: %s", workerID, msg.ID)

		case <-ctx.Done():
			// Context 取消，丢弃消息并退出
			s.logger.Warnf(ctx, "[Subscriber-%d] Dropping message due to shutdown: %s", workerID, msg.ID)
			return
		}

		// 3. 速率控制 + 退出检查
		select {
		case <-ctx.Done():
			s.logger.Infof(ctx, "[Subscriber-%d] Context cancelled, exiting", workerID)
			return

		case <-time.After(s.cfg.Rate):
			// 速率限制通过，继续下一次循环
			continue
		}
	}
}
