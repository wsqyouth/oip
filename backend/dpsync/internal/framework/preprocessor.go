package framework

import (
	"context"
	"fmt"
)

// PreProcessor 函数链处理器
type PreProcessor struct {
	processFuncs []ProcessorFunc
}

// NewPreProcessor 创建函数链处理器
func NewPreProcessor(processFuncs []ProcessorFunc) *PreProcessor {
	return &PreProcessor{
		processFuncs: processFuncs,
	}
}

// Run 执行函数链
// 任一函数返回 error 则立即停止
func (p *PreProcessor) Run(ctx context.Context) error {
	for i, processFunc := range p.processFuncs {
		if err := processFunc(ctx); err != nil {
			return fmt.Errorf("processor[%d] failed: %w", i, err)
		}
	}
	return nil
}
