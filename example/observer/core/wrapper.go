package core

import (
	"context"
	"log"
	"sync"
	"time"
)

// observerWrapper 观察者包装器
// 负责管理每个观察者的 goroutine 生命周期、错误处理和统计
type observerWrapper struct {
	observer  Observer
	eventCh   chan Event
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup
	config    *EventBusConfig
	stats     *ObserverStats
	started   bool
	startOnce sync.Once
}

// newObserverWrapper 创建观察者包装器
func newObserverWrapper(observer Observer, parentCtx context.Context, config *EventBusConfig) *observerWrapper {
	ctx, cancel := context.WithCancel(parentCtx)

	bufferSize := config.BufferSize
	if bufferSize < 0 {
		bufferSize = 0
	}

	return &observerWrapper{
		observer: observer,
		eventCh:  make(chan Event, bufferSize),
		ctx:      ctx,
		cancel:   cancel,
		config:   config,
		stats:    &ObserverStats{},
	}
}

// start 启动观察者的 goroutine
func (w *observerWrapper) start() {
	w.startOnce.Do(func() {
		w.started = true
		w.wg.Add(1)
		go w.loop()
	})
}

// loop 观察者的主循环，监听事件并处理
func (w *observerWrapper) loop() {
	defer func() {
		w.wg.Done()
		log.Printf("Observer %s: goroutine 已退出", w.observer.ID())
	}()

	log.Printf("Observer %s: goroutine 已启动", w.observer.ID())

	for {
		select {
		case event := <-w.eventCh:
			w.handleEvent(event)
		case <-w.ctx.Done():
			log.Printf("Observer %s: 收到停止信号，开始清理", w.observer.ID())
			// 处理剩余事件（可选，根据需求决定）
			w.drainEvents()
			return
		}
	}
}

// handleEvent 处理单个事件
func (w *observerWrapper) handleEvent(event Event) {
	startTime := time.Now()

	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(w.ctx, w.config.Timeout)
	defer cancel()

	// 使用 recover 防止 panic 影响其他观察者
	defer func() {
		if r := recover(); r != nil {
			log.Printf("Observer %s panic: %v, event: %s", w.observer.ID(), r, event.Type())
			w.stats.IncPanicCount()
		}

		// 记录处理时间
		processTime := time.Since(startTime)
		w.stats.AddProcessTime(processTime)
	}()

	// 处理事件
	err := w.observer.Handle(ctx, event)
	if err != nil {
		if err == context.DeadlineExceeded {
			log.Printf("Observer %s: 处理事件超时, event: %s", w.observer.ID(), event.Type())
			w.stats.IncTimeoutCount()
		} else {
			log.Printf("Observer %s: 处理事件失败, event: %s, error: %v", w.observer.ID(), event.Type(), err)
			w.stats.IncErrorCount()
		}
	} else {
		w.stats.IncSuccessCount()
	}
}

// drainEvents 清空剩余事件（优雅关闭时使用）
func (w *observerWrapper) drainEvents() {
	for {
		select {
		case event := <-w.eventCh:
			// 尝试快速处理剩余事件
			w.handleEvent(event)
		default:
			// 没有更多事件，退出
			return
		}
	}
}

// send 发送事件到观察者的 channel（非阻塞）
func (w *observerWrapper) send(event Event) bool {
	select {
	case w.eventCh <- event:
		return true
	case <-w.ctx.Done():
		// 观察者已关闭
		return false
	default:
		// channel 已满（如果是有缓冲的）
		// 这里可以选择丢弃或阻塞，我们选择记录日志
		log.Printf("Observer %s: channel 已满，事件可能丢失, event: %s", w.observer.ID(), event.Type())
		return false
	}
}

// stop 停止观察者
func (w *observerWrapper) stop() {
	if !w.started {
		return
	}
	w.cancel()
}

// wait 等待观察者处理完成
func (w *observerWrapper) wait() {
	w.wg.Wait()
}

// getStats 获取统计信息
func (w *observerWrapper) getStats() ObserverStatsSnapshot {
	return w.stats.GetStats()
}
