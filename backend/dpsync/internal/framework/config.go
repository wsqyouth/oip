package framework

import "time"

// SubscriberConfig Subscriber 配置
type SubscriberConfig struct {
	QueueName    string        // 队列名称
	Concurrency  int           // 并发拉取数
	Timeout      time.Duration // 拉取超时
	TTR          time.Duration // Time-To-Run
	Rate         time.Duration // 速率限制（拉取间隔）
	ErrorBackoff time.Duration // 错误退避时间
}

// ProcessorConfig Processor 配置
type ProcessorConfig struct {
	Concurrency int           // 并发处理数
	BufferSize  int           // inputChan 缓冲区大小
	Timeout     time.Duration // 单个消息处理超时
}
