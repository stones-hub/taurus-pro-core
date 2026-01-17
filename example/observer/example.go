package observer

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/stones-hub/taurus-pro-core/example/observer/core"
	"github.com/stones-hub/taurus-pro-core/example/observer/events"
	"github.com/stones-hub/taurus-pro-core/example/observer/observers"
)

// ExampleBasicUsage 基础使用示例
func ExampleBasicUsage() {
	log.Println("=== 基础使用示例 ===")

	// 1. 创建事件总线
	ctx := context.Background()
	bus := core.NewEventBus(ctx,
		core.WithBufferSize(100),
		core.WithTimeout(5*time.Second),
		core.WithStats(true),
	)

	// 2. 创建观察者
	loggerObserver := observers.NewLoggerObserver("logger-1")
	metricsObserver := observers.NewMetricsObserver("metrics-1")

	// 3. 订阅事件
	bus.Subscribe(events.EventTypeUserLogin, loggerObserver)
	bus.Subscribe(events.EventTypeUserLogin, metricsObserver)
	bus.Subscribe(events.EventTypeUserLogout, loggerObserver)

	// 4. 发布事件
	loginEvent := events.NewUserLoginEvent("user-123", "192.168.1.1", "Mozilla/5.0")
	bus.Publish(loginEvent)

	logoutEvent := events.NewUserLogoutEvent("user-123", "session-456")
	bus.Publish(logoutEvent)

	// 5. 等待事件处理
	time.Sleep(1 * time.Second)

	// 6. 查看统计信息
	stats := bus.GetStats()
	log.Printf("事件总线统计: 发布=%d, 分发=%d, 观察者=%d",
		stats.PublishedCount, stats.DeliveredCount, stats.ObserverCount)

	// 7. 优雅关闭
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	bus.Shutdown(shutdownCtx)
}

// ExampleFunctionObserver 函数式观察者示例
func ExampleFunctionObserver() {
	log.Println("\n=== 函数式观察者示例 ===")

	ctx := context.Background()
	bus := core.NewEventBus(ctx)

	// 使用函数式观察者（更 Go 风格）
	bus.Subscribe(events.EventTypeUserRegister, core.ObserverFunc(func(ctx context.Context, event core.Event) error {
		if regEvent, ok := event.(*events.UserRegisterEvent); ok {
			log.Printf("函数式观察者: 欢迎新用户 %s (ID: %s)", regEvent.Username, regEvent.UserID)
		}
		return nil
	}))

	// 使用命名函数式观察者
	bus.Subscribe(events.EventTypeUserRegister, core.NewNamedObserverFunc("welcome-handler", func(ctx context.Context, event core.Event) error {
		if regEvent, ok := event.(*events.UserRegisterEvent); ok {
			log.Printf("命名函数式观察者: 发送欢迎邮件给 %s", regEvent.Email)
		}
		return nil
	}))

	// 发布事件
	registerEvent := events.NewUserRegisterEvent("user-456", "alice", "alice@example.com")
	bus.Publish(registerEvent)

	time.Sleep(500 * time.Millisecond)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	bus.Shutdown(shutdownCtx)
}

// ExampleMultipleObservers 多个观察者示例
func ExampleMultipleObservers() {
	log.Println("\n=== 多个观察者示例 ===")

	ctx := context.Background()
	bus := core.NewEventBus(ctx)

	// 创建多个不同类型的观察者
	loggerObserver := observers.NewLoggerObserver("logger")
	auditObserver := observers.NewAuditObserver("audit")
	metricsObserver := observers.NewMetricsObserver("metrics")
	notificationObserver := observers.NewNotificationObserver("notification", []string{"email", "sms"})

	// 订阅不同的事件
	bus.Subscribe(events.EventTypeUserLogin, loggerObserver)
	bus.Subscribe(events.EventTypeUserLogin, auditObserver)
	bus.Subscribe(events.EventTypeUserLogin, metricsObserver)
	bus.Subscribe(events.EventTypeUserLogin, notificationObserver)

	bus.Subscribe(events.EventTypeUserRegister, loggerObserver)
	bus.Subscribe(events.EventTypeUserRegister, auditObserver)
	bus.Subscribe(events.EventTypeUserRegister, metricsObserver)
	bus.Subscribe(events.EventTypeUserRegister, notificationObserver)

	// 发布多个事件
	for i := 1; i <= 3; i++ {
		userID := fmt.Sprintf("user-%d", i)
		loginEvent := events.NewUserLoginEvent(userID, "192.168.1.100", "Chrome/120.0")
		bus.Publish(loginEvent)
		time.Sleep(100 * time.Millisecond)
	}

	registerEvent := events.NewUserRegisterEvent("user-new", "bob", "bob@example.com")
	bus.Publish(registerEvent)

	// 等待处理
	time.Sleep(2 * time.Second)

	// 查看指标
	metrics := metricsObserver.GetMetrics()
	log.Printf("指标统计: 登录=%d, 注册=%d", metrics.LoginCount, metrics.RegisterCount)

	// 查看审计记录
	auditRecords := auditObserver.GetAuditRecords()
	log.Printf("审计记录数: %d", len(auditRecords))

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	bus.Shutdown(shutdownCtx)
}

// ExampleConfigChange 配置变更示例
func ExampleConfigChange() {
	log.Println("\n=== 配置变更示例 ===")

	ctx := context.Background()
	bus := core.NewEventBus(ctx)

	// 订阅配置变更事件
	bus.Subscribe(events.EventTypeConfigChange, core.NewNamedObserverFunc("config-watcher", func(ctx context.Context, event core.Event) error {
		if configEvent, ok := event.(*events.ConfigChangeEvent); ok {
			log.Printf("配置变更: %s 从 %v 变更为 %v",
				configEvent.Key, configEvent.OldValue, configEvent.NewValue)
		}
		return nil
	}))

	// 模拟配置变更
	bus.Publish(events.NewConfigChangeEvent("database.host", "localhost", "192.168.1.100"))
	bus.Publish(events.NewConfigChangeEvent("app.timeout", 30, 60))

	time.Sleep(500 * time.Millisecond)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	bus.Shutdown(shutdownCtx)
}

// ExampleErrorHandling 错误处理示例
func ExampleErrorHandling() {
	log.Println("\n=== 错误处理示例 ===")

	ctx := context.Background()
	bus := core.NewEventBus(ctx)

	// 创建一个会出错的观察者
	errorObserver := core.NewNamedObserverFunc("error-observer", func(ctx context.Context, event core.Event) error {
		log.Printf("错误观察者: 尝试处理事件 %s", event.Type())
		return fmt.Errorf("处理失败: 模拟错误")
	})

	// 创建一个正常的观察者
	normalObserver := core.NewNamedObserverFunc("normal-observer", func(ctx context.Context, event core.Event) error {
		log.Printf("正常观察者: 成功处理事件 %s", event.Type())
		return nil
	})

	// 订阅同一个事件
	bus.Subscribe(events.EventTypeUserLogin, errorObserver)
	bus.Subscribe(events.EventTypeUserLogin, normalObserver)

	// 发布事件
	loginEvent := events.NewUserLoginEvent("user-error", "192.168.1.1", "Test")
	bus.Publish(loginEvent)

	// 等待处理（注意：一个观察者失败不影响其他观察者）
	time.Sleep(1 * time.Second)

	// 查看观察者统计
	errorStats, _ := bus.GetObserverStats(events.EventTypeUserLogin, "error-observer")
	normalStats, _ := bus.GetObserverStats(events.EventTypeUserLogin, "normal-observer")

	log.Printf("错误观察者统计: 成功=%d, 失败=%d", errorStats.SuccessCount, errorStats.ErrorCount)
	log.Printf("正常观察者统计: 成功=%d, 失败=%d", normalStats.SuccessCount, normalStats.ErrorCount)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	bus.Shutdown(shutdownCtx)
}

// ExampleUnsubscribe 取消订阅示例
func ExampleUnsubscribe() {
	log.Println("\n=== 取消订阅示例 ===")

	ctx := context.Background()
	bus := core.NewEventBus(ctx)

	observer1 := observers.NewLoggerObserver("logger-1")
	observer2 := observers.NewLoggerObserver("logger-2")

	// 订阅
	bus.Subscribe(events.EventTypeUserLogin, observer1)
	bus.Subscribe(events.EventTypeUserLogin, observer2)

	// 发布事件（两个观察者都会收到）
	bus.Publish(events.NewUserLoginEvent("user-1", "192.168.1.1", "Chrome"))
	time.Sleep(500 * time.Millisecond)

	// 取消 observer1 的订阅
	bus.Unsubscribe(events.EventTypeUserLogin, observer1.ID())

	// 再次发布事件（只有 observer2 会收到）
	bus.Publish(events.NewUserLoginEvent("user-2", "192.168.1.1", "Chrome"))
	time.Sleep(500 * time.Millisecond)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	bus.Shutdown(shutdownCtx)
}
