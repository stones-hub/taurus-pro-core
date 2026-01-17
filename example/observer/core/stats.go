package core

import (
	"sync"
	"sync/atomic"
	"time"
)

// ObserverStats 观察者统计信息
type ObserverStats struct {
	// 成功处理的事件数
	SuccessCount int64
	// 处理失败的事件数
	ErrorCount int64
	// 超时的事件数
	TimeoutCount int64
	// Panic 次数
	PanicCount int64
	// 总处理时间
	TotalProcessTime time.Duration
	// 最后处理时间
	LastProcessTime time.Time
	mu              sync.RWMutex
}

// IncSuccessCount 增加成功计数
func (s *ObserverStats) IncSuccessCount() {
	atomic.AddInt64(&s.SuccessCount, 1)
}

// IncErrorCount 增加错误计数
func (s *ObserverStats) IncErrorCount() {
	atomic.AddInt64(&s.ErrorCount, 1)
}

// IncTimeoutCount 增加超时计数
func (s *ObserverStats) IncTimeoutCount() {
	atomic.AddInt64(&s.TimeoutCount, 1)
}

// IncPanicCount 增加 Panic 计数
func (s *ObserverStats) IncPanicCount() {
	atomic.AddInt64(&s.PanicCount, 1)
}

// AddProcessTime 增加处理时间
func (s *ObserverStats) AddProcessTime(duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.TotalProcessTime += duration
	s.LastProcessTime = time.Now()
}

// GetStats 获取统计信息快照
func (s *ObserverStats) GetStats() ObserverStatsSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return ObserverStatsSnapshot{
		SuccessCount:     atomic.LoadInt64(&s.SuccessCount),
		ErrorCount:       atomic.LoadInt64(&s.ErrorCount),
		TimeoutCount:     atomic.LoadInt64(&s.TimeoutCount),
		PanicCount:       atomic.LoadInt64(&s.PanicCount),
		TotalProcessTime: s.TotalProcessTime,
		LastProcessTime:  s.LastProcessTime,
	}
}

// ObserverStatsSnapshot 观察者统计信息快照
type ObserverStatsSnapshot struct {
	SuccessCount     int64
	ErrorCount       int64
	TimeoutCount     int64
	PanicCount       int64
	TotalProcessTime time.Duration
	LastProcessTime  time.Time
}

// EventBusStats 事件总线统计信息
type EventBusStats struct {
	// 发布的事件总数
	PublishedCount int64
	// 成功分发的事件数
	DeliveredCount int64
	// 分发失败的事件数
	DeliveryFailedCount int64
	// 观察者数量
	ObserverCount int64
	// 订阅的事件类型数量
	EventTypeCount int64
	mu             sync.RWMutex
}

// IncPublishedCount 增加发布计数
func (s *EventBusStats) IncPublishedCount() {
	atomic.AddInt64(&s.PublishedCount, 1)
}

// IncDeliveredCount 增加分发计数
func (s *EventBusStats) IncDeliveredCount() {
	atomic.AddInt64(&s.DeliveredCount, 1)
}

// IncDeliveryFailedCount 增加分发失败计数
func (s *EventBusStats) IncDeliveryFailedCount() {
	atomic.AddInt64(&s.DeliveryFailedCount, 1)
}

// SetObserverCount 设置观察者数量
func (s *EventBusStats) SetObserverCount(count int) {
	atomic.StoreInt64(&s.ObserverCount, int64(count))
}

// SetEventTypeCount 设置事件类型数量
func (s *EventBusStats) SetEventTypeCount(count int) {
	atomic.StoreInt64(&s.EventTypeCount, int64(count))
}

// GetStats 获取统计信息快照
func (s *EventBusStats) GetStats() EventBusStatsSnapshot {
	return EventBusStatsSnapshot{
		PublishedCount:      atomic.LoadInt64(&s.PublishedCount),
		DeliveredCount:      atomic.LoadInt64(&s.DeliveredCount),
		DeliveryFailedCount: atomic.LoadInt64(&s.DeliveryFailedCount),
		ObserverCount:       atomic.LoadInt64(&s.ObserverCount),
		EventTypeCount:      atomic.LoadInt64(&s.EventTypeCount),
	}
}

// EventBusStatsSnapshot 事件总线统计信息快照
type EventBusStatsSnapshot struct {
	PublishedCount      int64
	DeliveredCount      int64
	DeliveryFailedCount int64
	ObserverCount       int64
	EventTypeCount      int64
}
