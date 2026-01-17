package core

import "time"

// EventType 事件类型，用于类型安全的事件分类
type EventType string

// Event 事件接口，所有事件必须实现此接口
type Event interface {
	// Type 返回事件类型
	Type() EventType
	// Timestamp 返回事件发生时间
	Timestamp() time.Time
}

// BaseEvent 基础事件，提供通用字段
type BaseEvent struct {
	eventType  EventType
	timestamp  time.Time
}

// NewBaseEvent 创建基础事件
func NewBaseEvent(eventType EventType) BaseEvent {
	return BaseEvent{
		eventType: eventType,
		timestamp: time.Now(),
	}
}

// Type 返回事件类型
func (e BaseEvent) Type() EventType {
	return e.eventType
}

// Timestamp 返回事件时间戳
func (e BaseEvent) Timestamp() time.Time {
	return e.timestamp
}
