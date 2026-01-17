package core

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

var (
	// ErrEventBusClosed 事件总线已关闭
	ErrEventBusClosed = fmt.Errorf("event bus is closed")
	// ErrObserverNotFound 观察者未找到
	ErrObserverNotFound = fmt.Errorf("observer not found")
)

// EventBus 事件总线
// 负责管理事件订阅和分发
type EventBus struct {
	// 生命周期管理
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// 订阅关系：EventType -> []*observerWrapper
	subscriptions map[EventType][]*observerWrapper
	mu            sync.RWMutex

	// 配置
	config *EventBusConfig

	// 统计信息
	stats *EventBusStats

	// 优雅关闭
	shutdownOnce sync.Once
	shutdownCh   chan struct{}
	closed       bool
	closedMu     sync.RWMutex
}

// NewEventBus 创建新的事件总线
func NewEventBus(ctx context.Context, opts ...EventBusOption) *EventBus {
	busCtx, cancel := context.WithCancel(ctx)

	config := DefaultEventBusConfig()
	// 应用选项
	for _, opt := range opts {
		opt(config)
	}

	eb := &EventBus{
		ctx:           busCtx,
		cancel:        cancel,
		subscriptions: make(map[EventType][]*observerWrapper),
		config:        config,
		stats:         &EventBusStats{},
		shutdownCh:    make(chan struct{}),
	}

	return eb
}

// EventBusOption 事件总线配置选项
type EventBusOption func(*EventBusConfig)

// WithBufferSize 设置缓冲区大小
func WithBufferSize(size int) EventBusOption {
	return func(c *EventBusConfig) {
		c.BufferSize = size
	}
}

// WithTimeout 设置超时时间
func WithTimeout(timeout time.Duration) EventBusOption {
	return func(c *EventBusConfig) {
		c.Timeout = timeout
	}
}

// WithStats 设置是否启用统计
func WithStats(enable bool) EventBusOption {
	return func(c *EventBusConfig) {
		c.EnableStats = enable
	}
}

// Subscribe 订阅事件
// eventType: 要订阅的事件类型
// observer: 观察者实现
func (eb *EventBus) Subscribe(eventType EventType, observer Observer) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	// 检查是否已关闭
	if eb.isClosed() {
		return ErrEventBusClosed
	}

	// 创建观察者包装器
	wrapper := newObserverWrapper(observer, eb.ctx, eb.config)

	// 添加到订阅列表
	eb.subscriptions[eventType] = append(eb.subscriptions[eventType], wrapper)

	// 启动观察者 goroutine
	wrapper.start()

	// 更新统计
	if eb.config.EnableStats {
		eb.updateStats()
	}

	log.Printf("Observer %s 已订阅事件类型: %s", observer.ID(), eventType)
	return nil
}

// Unsubscribe 取消订阅
func (eb *EventBus) Unsubscribe(eventType EventType, observerID string) error {
	eb.mu.Lock()
	defer eb.mu.Unlock()

	if eb.isClosed() {
		return ErrEventBusClosed
	}

	observers, exists := eb.subscriptions[eventType]
	if !exists {
		return ErrObserverNotFound
	}

	// 查找并移除观察者
	for i, wrapper := range observers {
		if wrapper.observer.ID() == observerID {
			// 停止观察者
			wrapper.stop()

			// 从列表中移除
			eb.subscriptions[eventType] = append(observers[:i], observers[i+1:]...)

			// 如果该事件类型没有观察者了，删除映射
			if len(eb.subscriptions[eventType]) == 0 {
				delete(eb.subscriptions, eventType)
			}

			// 更新统计
			if eb.config.EnableStats {
				eb.updateStats()
			}

			log.Printf("Observer %s 已取消订阅事件类型: %s", observerID, eventType)
			return nil
		}
	}

	return ErrObserverNotFound
}

// Publish 发布事件（非阻塞）
func (eb *EventBus) Publish(event Event) error {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	if eb.isClosed() {
		return ErrEventBusClosed
	}

	// 更新发布统计
	if eb.config.EnableStats {
		eb.stats.IncPublishedCount()
	}

	// 获取该事件类型的所有观察者
	observers := eb.subscriptions[event.Type()]
	if len(observers) == 0 {
		log.Printf("事件类型 %s 没有观察者", event.Type())
		return nil
	}

	// 为每个观察者发送事件（非阻塞）
	deliveredCount := 0
	for _, wrapper := range observers {
		if wrapper.send(event) {
			deliveredCount++
			if eb.config.EnableStats {
				eb.stats.IncDeliveredCount()
			}
		} else {
			if eb.config.EnableStats {
				eb.stats.IncDeliveryFailedCount()
			}
		}
	}

	log.Printf("事件 %s 已发布，分发给 %d/%d 个观察者", event.Type(), deliveredCount, len(observers))
	return nil
}

// Shutdown 优雅关闭事件总线
func (eb *EventBus) Shutdown(ctx context.Context) error {
	var shutdownErr error

	eb.shutdownOnce.Do(func() {
		log.Println("开始关闭事件总线...")

		// 1. 标记为关闭，停止接受新订阅
		eb.mu.Lock()
		eb.closed = true
		eb.mu.Unlock()
		close(eb.shutdownCh)

		// 2. 取消所有观察者的 context
		eb.cancel()

		// 3. 等待所有观察者处理完当前事件
		done := make(chan struct{})
		go func() {
			eb.mu.RLock()
			observers := make([]*observerWrapper, 0)
			for _, obsList := range eb.subscriptions {
				observers = append(observers, obsList...)
			}
			eb.mu.RUnlock()

			// 等待所有观察者完成
			for _, wrapper := range observers {
				wrapper.wait()
			}
			close(done)
		}()

		// 4. 带超时等待
		select {
		case <-done:
			log.Println("事件总线已优雅关闭")
		case <-ctx.Done():
			shutdownErr = ctx.Err()
			log.Printf("事件总线关闭超时: %v", shutdownErr)
		}
	})

	return shutdownErr
}

// GetStats 获取事件总线统计信息
func (eb *EventBus) GetStats() EventBusStatsSnapshot {
	if !eb.config.EnableStats {
		return EventBusStatsSnapshot{}
	}
	return eb.stats.GetStats()
}

// GetObserverStats 获取指定观察者的统计信息
func (eb *EventBus) GetObserverStats(eventType EventType, observerID string) (ObserverStatsSnapshot, error) {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	observers, exists := eb.subscriptions[eventType]
	if !exists {
		return ObserverStatsSnapshot{}, ErrObserverNotFound
	}

	for _, wrapper := range observers {
		if wrapper.observer.ID() == observerID {
			return wrapper.getStats(), nil
		}
	}

	return ObserverStatsSnapshot{}, ErrObserverNotFound
}

// isClosed 检查是否已关闭
func (eb *EventBus) isClosed() bool {
	eb.closedMu.RLock()
	defer eb.closedMu.RUnlock()
	return eb.closed
}

// updateStats 更新统计信息
func (eb *EventBus) updateStats() {
	totalObservers := 0
	for _, observers := range eb.subscriptions {
		totalObservers += len(observers)
	}
	eb.stats.SetObserverCount(totalObservers)
	eb.stats.SetEventTypeCount(len(eb.subscriptions))
}

// GetSubscriptions 获取所有订阅信息（用于调试）
func (eb *EventBus) GetSubscriptions() map[EventType][]string {
	eb.mu.RLock()
	defer eb.mu.RUnlock()

	result := make(map[EventType][]string)
	for eventType, observers := range eb.subscriptions {
		ids := make([]string, len(observers))
		for i, wrapper := range observers {
			ids[i] = wrapper.observer.ID()
		}
		result[eventType] = ids
	}
	return result
}
