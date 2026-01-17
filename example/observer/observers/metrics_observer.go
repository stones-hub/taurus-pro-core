package observers

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/stones-hub/taurus-pro-core/example/observer/core"
	"github.com/stones-hub/taurus-pro-core/example/observer/events"
)

// MetricsObserver 指标观察者
// 负责收集和统计事件指标
type MetricsObserver struct {
	id string
	mu sync.RWMutex

	// 指标数据
	loginCount    int64
	logoutCount   int64
	registerCount int64
	updateCount   int64
	configChangeCount int64

	// 时间窗口统计
	lastResetTime time.Time
}

// NewMetricsObserver 创建指标观察者
func NewMetricsObserver(id string) *MetricsObserver {
	return &MetricsObserver{
		id:            id,
		lastResetTime: time.Now(),
	}
}

// ID 返回观察者ID
func (o *MetricsObserver) ID() string {
	return o.id
}

// Handle 处理事件
func (o *MetricsObserver) Handle(ctx context.Context, event core.Event) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	switch event.Type() {
	case events.EventTypeUserLogin:
		o.loginCount++
		log.Printf("[Metrics] 登录事件计数: %d", o.loginCount)
	case events.EventTypeUserLogout:
		o.logoutCount++
		log.Printf("[Metrics] 登出事件计数: %d", o.logoutCount)
	case events.EventTypeUserRegister:
		o.registerCount++
		log.Printf("[Metrics] 注册事件计数: %d", o.registerCount)
	case events.EventTypeUserUpdate:
		o.updateCount++
		log.Printf("[Metrics] 更新事件计数: %d", o.updateCount)
	case events.EventTypeConfigChange:
		o.configChangeCount++
		log.Printf("[Metrics] 配置变更计数: %d", o.configChangeCount)
	}

	return nil
}

// GetMetrics 获取指标快照
func (o *MetricsObserver) GetMetrics() MetricsSnapshot {
	o.mu.RLock()
	defer o.mu.RUnlock()

	return MetricsSnapshot{
		LoginCount:        o.loginCount,
		LogoutCount:       o.logoutCount,
		RegisterCount:     o.registerCount,
		UpdateCount:       o.updateCount,
		ConfigChangeCount: o.configChangeCount,
		LastResetTime:     o.lastResetTime,
	}
}

// Reset 重置指标
func (o *MetricsObserver) Reset() {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.loginCount = 0
	o.logoutCount = 0
	o.registerCount = 0
	o.updateCount = 0
	o.configChangeCount = 0
	o.lastResetTime = time.Now()
}

// MetricsSnapshot 指标快照
type MetricsSnapshot struct {
	LoginCount        int64
	LogoutCount       int64
	RegisterCount     int64
	UpdateCount       int64
	ConfigChangeCount int64
	LastResetTime     time.Time
}
