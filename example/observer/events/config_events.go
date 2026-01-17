package events

import (
	"time"

	"github.com/stones-hub/taurus-pro-core/example/observer/core"
)

// 配置相关事件类型
const (
	EventTypeConfigChange core.EventType = "config.change"
)

// ConfigChangeEvent 配置变更事件
type ConfigChangeEvent struct {
	core.BaseEvent
	Key      string
	OldValue interface{}
	NewValue interface{}
	ChangeTime time.Time
}

// NewConfigChangeEvent 创建配置变更事件
func NewConfigChangeEvent(key string, oldValue, newValue interface{}) *ConfigChangeEvent {
	return &ConfigChangeEvent{
		BaseEvent:  core.NewBaseEvent(EventTypeConfigChange),
		Key:        key,
		OldValue:   oldValue,
		NewValue:   newValue,
		ChangeTime: time.Now(),
	}
}
