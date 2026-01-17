package service

// ============================================================================
// 服务层 (Service Layer)
// ============================================================================
// 策略模式的调用方，依赖 Notifier 接口而非具体实现
// 通过注册表获取策略，实现策略的动态选择和调用
// ============================================================================

import (
	"context"
	"fmt"
	"sync"

	"github.com/stones-hub/taurus-pro-core/example/strategy/core"
)

// NotificationService 通知服务
// 这是策略模式的调用方，它依赖 Notifier 接口，而不是具体实现
type NotificationService struct {
	registry *core.NotifierRegistry
	// 可以添加其他依赖，如日志、监控等
}

// NewNotificationService 创建通知服务
func NewNotificationService(registry *core.NotifierRegistry) *NotificationService {
	if registry == nil {
		registry = core.NewNotifierRegistry()
	}
	return &NotificationService{
		registry: registry,
	}
}

// SendNotification 发送通知（单一策略）
// 这是策略模式的核心调用：通过接口调用，不依赖具体实现
func (s *NotificationService) SendNotification(ctx context.Context, strategyName string, notification *core.Notification) error {
	// 1. 获取策略（通过名称）
	notifier, err := s.registry.Get(strategyName)
	if err != nil {
		return fmt.Errorf("failed to get notifier: %w", err)
	}

	// 2. 使用策略发送通知
	// 这里只依赖 Notifier 接口，不知道具体是哪个实现
	return notifier.Send(ctx, notification)
}

// SendNotificationBatch 批量发送通知（多个策略）
// 演示如何同时使用多个策略
func (s *NotificationService) SendNotificationBatch(ctx context.Context, strategyNames []string, notification *core.Notification) map[string]error {
	results := make(map[string]error)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, name := range strategyNames {
		wg.Add(1)
		go func(strategyName string) {
			defer wg.Done()
			err := s.SendNotification(ctx, strategyName, notification)
			mu.Lock()
			results[strategyName] = err
			mu.Unlock()
		}(name)
	}

	wg.Wait()
	return results
}

// SendNotificationWithFallback 带降级策略的发送
// 如果主策略失败，自动使用备用策略
func (s *NotificationService) SendNotificationWithFallback(ctx context.Context, primaryStrategy, fallbackStrategy string, notification *core.Notification) error {
	// 先尝试主策略
	err := s.SendNotification(ctx, primaryStrategy, notification)
	if err == nil {
		return nil
	}

	// 主策略失败，使用备用策略
	fmt.Printf("Primary strategy '%s' failed: %v, trying fallback '%s'\n", primaryStrategy, err, fallbackStrategy)
	return s.SendNotification(ctx, fallbackStrategy, notification)
}

// ListAvailableStrategies 列出所有可用的策略
func (s *NotificationService) ListAvailableStrategies() []string {
	return s.registry.List()
}

// ValidateNotification 验证通知（使用指定策略的验证逻辑）
func (s *NotificationService) ValidateNotification(strategyName string, notification *core.Notification) error {
	notifier, err := s.registry.Get(strategyName)
	if err != nil {
		return fmt.Errorf("failed to get notifier: %w", err)
	}
	return notifier.Validate(notification)
}
