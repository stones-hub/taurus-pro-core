package core

import "time"

// EventBusConfig 事件总线配置
type EventBusConfig struct {
	// BufferSize 每个观察者的 channel 缓冲区大小
	// 0 表示无缓冲 channel，>0 表示缓冲 channel
	BufferSize int
	// Timeout 观察者处理事件的超时时间
	Timeout time.Duration
	// EnableStats 是否启用统计信息
	EnableStats bool
}

// DefaultEventBusConfig 返回默认配置
func DefaultEventBusConfig() *EventBusConfig {
	return &EventBusConfig{
		BufferSize:  100,
		Timeout:     5 * time.Second,
		EnableStats: true,
	}
}

// WithBufferSize 设置缓冲区大小
func (c *EventBusConfig) WithBufferSize(size int) *EventBusConfig {
	c.BufferSize = size
	return c
}

// WithTimeout 设置超时时间
func (c *EventBusConfig) WithTimeout(timeout time.Duration) *EventBusConfig {
	c.Timeout = timeout
	return c
}

// WithStats 设置是否启用统计
func (c *EventBusConfig) WithStats(enable bool) *EventBusConfig {
	c.EnableStats = enable
	return c
}
