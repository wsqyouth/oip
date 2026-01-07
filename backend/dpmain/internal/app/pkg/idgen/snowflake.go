package idgen

import (
	"sync"
	"time"
)

// SnowflakeIDGenerator 简化的雪花ID生成器
// ID格式: 时间戳(10位) + 机器ID(2位) + 序列号(3位) = 15位数字
// 为分库分表预留扩展空间
type SnowflakeIDGenerator struct {
	mu        sync.Mutex
	epoch     int64 // 起始时间戳 (2024-01-01 00:00:00)
	machineID int64 // 机器ID (0-99)
	sequence  int64 // 序列号 (0-999)
	lastTime  int64 // 上次生成ID的时间戳
}

const (
	machineBits  = 2   // 机器ID位数，支持100个实例
	sequenceBits = 3   // 序列号位数，每毫秒支持1000个ID
	maxMachineID = 99  // 最大机器ID
	maxSequence  = 999 // 最大序列号
)

// NewSnowflakeIDGenerator 创建ID生成器
// machineID: 机器ID，范围 0-99
func NewSnowflakeIDGenerator(machineID int64) *SnowflakeIDGenerator {
	if machineID < 0 || machineID > maxMachineID {
		machineID = 0
	}

	// 使用 2024-01-01 00:00:00 作为起始时间
	epoch := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Unix()

	return &SnowflakeIDGenerator{
		epoch:     epoch,
		machineID: machineID,
		sequence:  0,
		lastTime:  0,
	}
}

// NextID 生成下一个ID
func (g *SnowflakeIDGenerator) NextID() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().Unix()

	if now == g.lastTime {
		// 同一秒内，序列号递增
		g.sequence = (g.sequence + 1) % (maxSequence + 1)
		if g.sequence == 0 {
			// 序列号用尽，等待下一秒
			for now <= g.lastTime {
				now = time.Now().Unix()
			}
		}
	} else {
		// 新的一秒，重置序列号
		g.sequence = 0
	}

	g.lastTime = now

	// 计算时间偏移（从epoch开始的秒数）
	timestamp := now - g.epoch

	// 组合ID: 时间戳(10位) * 100000 + 机器ID(2位) * 1000 + 序列号(3位)
	// 例如: 12345 * 100000 + 01 * 1000 + 234 = 1234501234
	id := timestamp*100000 + g.machineID*1000 + g.sequence

	return id
}

// 全局默认ID生成器（机器ID为1）
var defaultGenerator = NewSnowflakeIDGenerator(1)

// GenerateID 生成ID（使用默认生成器）
func GenerateID() int64 {
	return defaultGenerator.NextID()
}
