package strategies

// ============================================================================
// 策略层 (Strategy Layer) - 短信策略实现
// ============================================================================

import (
	"context"
	"fmt"
	"time"

	"esim/test/strategy/core"
)

// SMSNotifier 短信通知策略实现
type SMSNotifier struct {
	apiKey    string
	apiSecret string
	// 可以添加更多配置，如服务商信息等
}

// NewSMSNotifier 创建短信通知策略
func NewSMSNotifier(apiKey, apiSecret string) *SMSNotifier {
	return &SMSNotifier{
		apiKey:    apiKey,
		apiSecret: apiSecret,
	}
}

// Name 返回策略名称
func (n *SMSNotifier) Name() string {
	return "sms"
}

// Validate 验证短信通知参数
func (n *SMSNotifier) Validate(notification *core.Notification) error {
	if notification == nil {
		return fmt.Errorf("notification cannot be nil")
	}
	if notification.To == "" {
		return fmt.Errorf("phone number is required")
	}
	// 简单的手机号验证（实际项目中应该更严格）
	if len(notification.To) < 10 {
		return fmt.Errorf("invalid phone number: %s", notification.To)
	}
	if notification.Content == "" {
		return fmt.Errorf("sms content is required")
	}
	// 短信内容长度限制
	if len(notification.Content) > 500 {
		return fmt.Errorf("sms content too long, max 500 characters")
	}
	return nil
}

// Send 发送短信通知
func (n *SMSNotifier) Send(ctx context.Context, notification *core.Notification) error {
	// 先验证
	if err := n.Validate(notification); err != nil {
		return fmt.Errorf("sms validation failed: %w", err)
	}

	// 模拟发送短信的过程
	// 在实际项目中，这里会调用真实的短信服务商 API
	keyPreview := n.apiKey
	if len(keyPreview) > 8 {
		keyPreview = keyPreview[:8]
	}
	fmt.Printf("[SMSNotifier] Sending SMS via API (key: %s...)\n", keyPreview)
	fmt.Printf("  Phone: %s\n", notification.To)
	fmt.Printf("  Content: %s\n", notification.Content)

	// 模拟网络延迟（短信通常比邮件快）
	select {
	case <-time.After(50 * time.Millisecond):
		// 模拟发送成功
		fmt.Printf("[SMSNotifier] SMS sent successfully\n")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("sms send cancelled: %w", ctx.Err())
	}
}
