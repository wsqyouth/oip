package framework

import "time"

// Message 消息结构（框架内部流转）
type Message struct {
	ID       string                 // 消息 ID
	Queue    string                 // 队列名称
	Data     []byte                 // 原始 Job 数据
	Attempts int                    // 重试次数
	Extra    map[string]interface{} // 扩展字段
}

// ProcessResult 处理结果
type ProcessResult struct {
	Success bool          // 是否成功
	Error   error         // 错误信息
	RetryIn time.Duration // 重试延迟
}
