package strategies

// ============================================================================
// 策略层 (Strategy Layer) - 邮件策略实现
// ============================================================================
// 实现 Notifier 接口的具体策略，每个策略独立实现，互不依赖
// ============================================================================

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/stones-hub/taurus-pro-core/example/strategy/core"
)

// EmailNotifier 邮件通知策略实现
type EmailNotifier struct {
	smtpHost string
	smtpPort int
	// 可以添加更多配置，如认证信息等
}

// NewEmailNotifier 创建邮件通知策略
func NewEmailNotifier(smtpHost string, smtpPort int) *EmailNotifier {
	return &EmailNotifier{
		smtpHost: smtpHost,
		smtpPort: smtpPort,
	}
}

// Name 返回策略名称
func (n *EmailNotifier) Name() string {
	return "email"
}

// Validate 验证邮件通知参数
func (n *EmailNotifier) Validate(notification *core.Notification) error {
	if notification == nil {
		return fmt.Errorf("notification cannot be nil")
	}
	if notification.To == "" {
		return fmt.Errorf("email address is required")
	}
	// 简单的邮箱格式验证
	if !strings.Contains(notification.To, "@") {
		return fmt.Errorf("invalid email address: %s", notification.To)
	}
	if notification.Title == "" {
		return fmt.Errorf("email title is required")
	}
	if notification.Content == "" {
		return fmt.Errorf("email content is required")
	}
	return nil
}

// Send 发送邮件通知
func (n *EmailNotifier) Send(ctx context.Context, notification *core.Notification) error {
	// 先验证
	if err := n.Validate(notification); err != nil {
		return fmt.Errorf("email validation failed: %w", err)
	}

	// 模拟发送邮件的过程
	// 在实际项目中，这里会调用真实的 SMTP 服务
	fmt.Printf("[EmailNotifier] Sending email via %s:%d\n", n.smtpHost, n.smtpPort)
	fmt.Printf("  To: %s\n", notification.To)
	fmt.Printf("  Subject: %s\n", notification.Title)
	fmt.Printf("  Body: %s\n", notification.Content)

	// 模拟网络延迟
	select {
	case <-time.After(100 * time.Millisecond):
		// 模拟发送成功
		fmt.Printf("[EmailNotifier] Email sent successfully\n")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("email send cancelled: %w", ctx.Err())
	}
}
