package strategy

import (
	"context"
	"testing"
	"time"

	"esim/test/strategy/core"
	"esim/test/strategy/service"
	"esim/test/strategy/strategies"
)

// TestNotifierRegistry 测试策略注册表
func TestNotifierRegistry(t *testing.T) {
	registry := core.NewNotifierRegistry()

	// 测试注册
	emailNotifier := strategies.NewEmailNotifier("smtp.test.com", 587)
	registry.Register(emailNotifier)

	// 测试获取
	notifier, err := registry.Get("email")
	if err != nil {
		t.Fatalf("Failed to get email notifier: %v", err)
	}
	if notifier.Name() != "email" {
		t.Errorf("Expected name 'email', got '%s'", notifier.Name())
	}

	// 测试获取不存在的策略
	_, err = registry.Get("nonexistent")
	if err == nil {
		t.Error("Expected error when getting nonexistent strategy")
	}

	// 测试列出所有策略
	registry.Register(strategies.NewSMSNotifier("key", "secret"))
	strategyList := registry.List()
	if len(strategyList) != 2 {
		t.Errorf("Expected 2 strategies, got %d", len(strategyList))
	}

	// 测试检查存在性
	if !registry.Exists("email") {
		t.Error("Expected email strategy to exist")
	}
	if registry.Exists("nonexistent") {
		t.Error("Expected nonexistent strategy to not exist")
	}
}

// TestEmailNotifier 测试邮件通知策略
func TestEmailNotifier(t *testing.T) {
	notifier := strategies.NewEmailNotifier("smtp.test.com", 587)

	// 测试名称
	if notifier.Name() != "email" {
		t.Errorf("Expected name 'email', got '%s'", notifier.Name())
	}

	// 测试验证 - 有效通知
	validNotification := &core.Notification{
		Title:   "Test",
		Content: "Test content",
		To:      "test@example.com",
	}
	if err := notifier.Validate(validNotification); err != nil {
		t.Errorf("Expected valid notification, got error: %v", err)
	}

	// 测试验证 - 无效邮箱
	invalidNotification := &core.Notification{
		Title:   "Test",
		Content: "Test content",
		To:      "invalid-email",
	}
	if err := notifier.Validate(invalidNotification); err == nil {
		t.Error("Expected validation error for invalid email")
	}

	// 测试发送
	ctx := context.Background()
	if err := notifier.Send(ctx, validNotification); err != nil {
		t.Errorf("Expected successful send, got error: %v", err)
	}
}

// TestSMSNotifier 测试短信通知策略
func TestSMSNotifier(t *testing.T) {
	notifier := strategies.NewSMSNotifier("test_key", "test_secret")

	if notifier.Name() != "sms" {
		t.Errorf("Expected name 'sms', got '%s'", notifier.Name())
	}

	// 测试验证 - 内容过长
	longContent := make([]byte, 501)
	for i := range longContent {
		longContent[i] = 'a'
	}
	invalidNotification := &core.Notification{
		Content: string(longContent),
		To:      "13800138000",
	}
	if err := notifier.Validate(invalidNotification); err == nil {
		t.Error("Expected validation error for content too long")
	}
}

// TestWechatNotifier 测试微信通知策略
func TestWechatNotifier(t *testing.T) {
	notifier := strategies.NewWechatNotifier("test_app_id", "test_app_secret")

	if notifier.Name() != "wechat" {
		t.Errorf("Expected name 'wechat', got '%s'", notifier.Name())
	}

	// 测试验证 - OpenID 太短
	invalidNotification := &core.Notification{
		Title:   "Test",
		Content: "Test",
		To:      "short",
	}
	if err := notifier.Validate(invalidNotification); err == nil {
		t.Error("Expected validation error for short openid")
	}
}

// TestNotificationService 测试通知服务
func TestNotificationService(t *testing.T) {
	registry := core.NewNotifierRegistry()
	registry.Register(strategies.NewEmailNotifier("smtp.test.com", 587))
	registry.Register(strategies.NewSMSNotifier("key", "secret"))

	svc := service.NewNotificationService(registry)

	ctx := context.Background()
	notification := &core.Notification{
		Title:   "Test",
		Content: "Test content",
		To:      "test@example.com",
	}

	// 测试发送
	if err := svc.SendNotification(ctx, "email", notification); err != nil {
		t.Errorf("Expected successful send, got error: %v", err)
	}

	// 测试不存在的策略
	if err := svc.SendNotification(ctx, "nonexistent", notification); err == nil {
		t.Error("Expected error for nonexistent strategy")
	}
}

// TestNotificationServiceBatch 测试批量发送
func TestNotificationServiceBatch(t *testing.T) {
	registry := core.NewNotifierRegistry()
	registry.Register(strategies.NewEmailNotifier("smtp.test.com", 587))
	registry.Register(strategies.NewSMSNotifier("key", "secret"))

	svc := service.NewNotificationService(registry)

	ctx := context.Background()
	notification := &core.Notification{
		Title:   "Test",
		Content: "Test content",
		To:      "test@example.com",
	}

	results := svc.SendNotificationBatch(ctx, []string{"email", "sms"}, notification)
	if len(results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(results))
	}

	for strategy, err := range results {
		if err != nil {
			t.Errorf("Strategy '%s' failed: %v", strategy, err)
		}
	}
}

// TestNotificationServiceFallback 测试降级策略
func TestNotificationServiceFallback(t *testing.T) {
	registry := core.NewNotifierRegistry()
	registry.Register(strategies.NewEmailNotifier("smtp.test.com", 587))
	registry.Register(strategies.NewSMSNotifier("key", "secret"))

	svc := service.NewNotificationService(registry)

	ctx := context.Background()
	notification := &core.Notification{
		Title:   "Test",
		Content: "Test content",
		To:      "13800138000",
	}

	// 使用有效的策略，应该成功
	if err := svc.SendNotificationWithFallback(ctx, "email", "sms", notification); err != nil {
		t.Errorf("Expected successful send with fallback, got error: %v", err)
	}
}

// TestContextCancellation 测试 Context 取消
func TestContextCancellation(t *testing.T) {
	registry := core.NewNotifierRegistry()
	registry.Register(strategies.NewEmailNotifier("smtp.test.com", 587))
	svc := service.NewNotificationService(registry)

	// 创建一个会立即取消的 context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	notification := &core.Notification{
		Title:   "Test",
		Content: "Test content",
		To:      "test@example.com",
	}

	// 应该因为 context 取消而失败
	err := svc.SendNotification(ctx, "email", notification)
	if err == nil {
		t.Error("Expected error due to context cancellation")
	}
}

// TestContextTimeout 测试 Context 超时
func TestContextTimeout(t *testing.T) {
	registry := core.NewNotifierRegistry()
	registry.Register(strategies.NewEmailNotifier("smtp.test.com", 587))
	svc := service.NewNotificationService(registry)

	// 创建一个很短的超时
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	notification := &core.Notification{
		Title:   "Test",
		Content: "Test content",
		To:      "test@example.com",
	}

	// 应该因为超时而失败
	err := svc.SendNotification(ctx, "email", notification)
	if err == nil {
		t.Error("Expected error due to context timeout")
	}
}

// BenchmarkNotificationService 性能基准测试
func BenchmarkNotificationService(b *testing.B) {
	registry := core.NewNotifierRegistry()
	registry.Register(strategies.NewEmailNotifier("smtp.test.com", 587))
	svc := service.NewNotificationService(registry)

	ctx := context.Background()
	notification := &core.Notification{
		Title:   "Benchmark",
		Content: "Benchmark content",
		To:      "test@example.com",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = svc.SendNotification(ctx, "email", notification)
	}
}
