package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/stones-hub/taurus-pro-core/example/observer/core"
	"github.com/stones-hub/taurus-pro-core/example/observer/events"
	"github.com/stones-hub/taurus-pro-core/example/observer/observers"
)

func main() {
	log.Println("=== Go 风格观察者模式演示 ===")

	// 创建根 context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建事件总线
	bus := core.NewEventBus(ctx,
		core.WithBufferSize(100),
		core.WithTimeout(5*time.Second),
		core.WithStats(true),
	)

	// 创建并注册观察者
	setupObservers(bus)

	// 启动事件发布模拟
	go simulateEvents(bus)

	// 定期打印统计信息
	go printStats(bus)

	// 等待中断信号
	waitForInterrupt()

	// 优雅关闭
	log.Println("开始优雅关闭...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := bus.Shutdown(shutdownCtx); err != nil {
		log.Printf("关闭超时: %v", err)
	} else {
		log.Println("事件总线已优雅关闭")
	}

	// 打印最终统计
	printFinalStats(bus)
}

// setupObservers 设置观察者
func setupObservers(bus *core.EventBus) {
	// 日志观察者
	loggerObserver := observers.NewLoggerObserver("logger")
	bus.Subscribe(events.EventTypeUserLogin, loggerObserver)
	bus.Subscribe(events.EventTypeUserLogout, loggerObserver)
	bus.Subscribe(events.EventTypeUserRegister, loggerObserver)
	bus.Subscribe(events.EventTypeUserUpdate, loggerObserver)

	// 审计观察者
	auditObserver := observers.NewAuditObserver("audit")
	bus.Subscribe(events.EventTypeUserLogin, auditObserver)
	bus.Subscribe(events.EventTypeUserLogout, auditObserver)
	bus.Subscribe(events.EventTypeUserRegister, auditObserver)

	// 指标观察者
	metricsObserver := observers.NewMetricsObserver("metrics")
	bus.Subscribe(events.EventTypeUserLogin, metricsObserver)
	bus.Subscribe(events.EventTypeUserLogout, metricsObserver)
	bus.Subscribe(events.EventTypeUserRegister, metricsObserver)
	bus.Subscribe(events.EventTypeUserUpdate, metricsObserver)
	bus.Subscribe(events.EventTypeConfigChange, metricsObserver)

	// 通知观察者
	notificationObserver := observers.NewNotificationObserver("notification", []string{"email", "sms"})
	bus.Subscribe(events.EventTypeUserRegister, notificationObserver)
	bus.Subscribe(events.EventTypeUserLogin, notificationObserver)

	// 配置变更观察者（函数式）
	bus.Subscribe(events.EventTypeConfigChange, core.NewNamedObserverFunc("config-watcher", func(ctx context.Context, event core.Event) error {
		if configEvent, ok := event.(*events.ConfigChangeEvent); ok {
			log.Printf("[ConfigWatcher] 配置 %s 已变更: %v -> %v",
				configEvent.Key, configEvent.OldValue, configEvent.NewValue)
		}
		return nil
	}))

	log.Println("所有观察者已注册")
}

// simulateEvents 模拟事件发布
func simulateEvents(bus *core.EventBus) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	userCounter := 1
	eventTypes := []func(int){
		func(id int) {
			event := events.NewUserLoginEvent(fmt.Sprintf("user-%d", id), "192.168.1.100", "Chrome/120.0")
			bus.Publish(event)
		},
		func(id int) {
			event := events.NewUserLogoutEvent(fmt.Sprintf("user-%d", id), fmt.Sprintf("session-%d", id))
			bus.Publish(event)
		},
		func(id int) {
			event := events.NewUserRegisterEvent(fmt.Sprintf("user-%d", id), fmt.Sprintf("user%d", id), fmt.Sprintf("user%d@example.com", id))
			bus.Publish(event)
		},
		func(id int) {
			event := events.NewUserUpdateEvent(fmt.Sprintf("user-%d", id), map[string]interface{}{
				"name": fmt.Sprintf("User %d", id),
			})
			bus.Publish(event)
		},
		func(id int) {
			event := events.NewConfigChangeEvent("app.timeout", 30, 60)
			bus.Publish(event)
		},
	}

	for {
		select {
		case <-ticker.C:
			// 随机选择一个事件类型发布
			eventType := eventTypes[userCounter%len(eventTypes)]
			eventType(userCounter)
			userCounter++
		}
	}
}

// printStats 定期打印统计信息
func printStats(bus *core.EventBus) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			stats := bus.GetStats()
			log.Printf("\n[统计] 发布=%d, 分发=%d, 失败=%d, 观察者=%d, 事件类型=%d\n",
				stats.PublishedCount,
				stats.DeliveredCount,
				stats.DeliveryFailedCount,
				stats.ObserverCount,
				stats.EventTypeCount,
			)

			// 打印订阅关系
			subscriptions := bus.GetSubscriptions()
			log.Println("[订阅关系]")
			for eventType, observerIDs := range subscriptions {
				log.Printf("  %s: %v", eventType, observerIDs)
			}
		}
	}
}

// printFinalStats 打印最终统计信息
func printFinalStats(bus *core.EventBus) {
	log.Println("\n=== 最终统计信息 ===")
	stats := bus.GetStats()
	log.Printf("总发布事件: %d", stats.PublishedCount)
	log.Printf("总分发事件: %d", stats.DeliveredCount)
	log.Printf("分发失败: %d", stats.DeliveryFailedCount)
	log.Printf("观察者数量: %d", stats.ObserverCount)
	log.Printf("事件类型数量: %d", stats.EventTypeCount)
}

// waitForInterrupt 等待中断信号
func waitForInterrupt() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan
	log.Println("\n收到中断信号，准备关闭...")
}
