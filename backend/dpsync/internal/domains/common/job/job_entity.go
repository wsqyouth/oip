package job

// Job 标准 Job 结构
type Job struct {
	Payload *JobPayload `json:"payload"`
}

// JobPayload Job 负载
type JobPayload struct {
	Data *JobPayloadData `json:"data"`
}

// JobPayloadData Job 数据
type JobPayloadData struct {
	// 元信息
	RequestID  string `json:"request_id"`  // 请求 ID（TraceID）
	OrgID      string `json:"org_id"`      // 组织 ID
	ActionType string `json:"action_type"` // 动作类型（路由键）
	ID         string `json:"id"`          // 业务 ID

	// 业务数据
	Data interface{} `json:"data"` // 具体业务数据

	// 扩展
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Meta 元数据
type Meta struct {
	RequestID  string // 请求 ID
	OrgID      string // 组织 ID
	ActionType string // 动作类型
	ID         string // 业务 ID
}
