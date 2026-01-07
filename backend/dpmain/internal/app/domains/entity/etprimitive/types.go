package etprimitive

// 基础类型和通用值对象（预留）

// Result 通用结果类型
type Result struct {
	Success bool
	Error   error
	Data    interface{}
}

// Pagination 分页参数
type Pagination struct {
	Page  int
	Limit int
	Total int64
}
