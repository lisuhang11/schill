package uuid

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Generator 基于 Snowflake 思想的分布式 ID 生成器
// 格式：42位时间戳(ms) + 10位机器ID + 12位序列号 = 64位
type Generator struct {
	mu        sync.Mutex
	machineID int64
	sequence  int64
	lastStamp int64
}

const (
	machineBits = 10
	sequenceBits = 12
	machineMax = -1 ^ (-1 << machineBits) // 1023
	sequenceMax = -1 ^ (-1 << sequenceBits) // 4095
	epoch = int64(1700000000000) // 2023-11-15 00:00:00 UTC in ms
)

var defaultGen *Generator

func init() {
	// 使用随机机器 ID（生产环境应从配置或 Etcd 获取）
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	defaultGen = &Generator{
		machineID: int64(r.Intn(int(machineMax + 1))),
	}
}

// NewGenerator 创建指定机器 ID 的生成器
func NewGenerator(machineID int64) (*Generator, error) {
	if machineID < 0 || machineID > machineMax {
		return nil, fmt.Errorf("machine id must be between 0 and %d", machineMax)
	}
	return &Generator{machineID: machineID}, nil
}

// NextID 生成下一个唯一 ID
func NextID() uint64 {
	return uint64(defaultGen.next())
}

// NextIDWithMachine 使用指定机器 ID 生成 ID
func NextIDWithMachine(machineID int64) (uint64, error) {
	g, err := NewGenerator(machineID)
	if err != nil {
		return 0, err
	}
	return uint64(g.next()), nil
}

func (g *Generator) next() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UnixMilli()
	if now < g.lastStamp {
		// 时钟回拨，等待追上
		for now < g.lastStamp {
			time.Sleep(time.Millisecond)
			now = time.Now().UnixMilli()
		}
	}

	if now == g.lastStamp {
		g.sequence = (g.sequence + 1) & sequenceMax
		if g.sequence == 0 {
			// 序列号用完，等待下一毫秒
			for now <= g.lastStamp {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		g.sequence = 0
	}

	g.lastStamp = now
	return ((now - epoch) << (machineBits + sequenceBits)) |
		(g.machineID << sequenceBits) |
		g.sequence
}
