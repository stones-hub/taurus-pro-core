package strategies

// ============================================================================
// 策略层 (Strategy Layer) - 微信策略实现
// ============================================================================

import (
	"context"
	"fmt"
	"time"

	"esim/test/strategy/core"
)

// WechatNotifier 微信通知策略实现
type WechatNotifier struct {
	appID     string
	appSecret string
	// 可以添加更多配置，如企业微信、公众号等
}

// NewWechatNotifier 创建微信通知策略
func NewWechatNotifier(appID, appSecret string) *WechatNotifier {
	return &WechatNotifier{
		appID:     appID,
		appSecret: appSecret,
	}
}

// Name 返回策略名称
func (n *WechatNotifier) Name() string {
	return "wechat"
}

// Validate 验证微信通知参数
func (n *WechatNotifier) Validate(notification *core.Notification) error {
	if notification == nil {
		return fmt.Errorf("notification cannot be nil")
	}
	if notification.To == "" {
		return fmt.Errorf("wechat openid is required")
	}
	// 微信 openid 通常是 28 位字符串
	if len(notification.To) < 20 {
		return fmt.Errorf("invalid wechat openid: %s", notification.To)
	}
	if notification.Title == "" && notification.Content == "" {
		return fmt.Errorf("wechat notification must have title or content")
	}
	return nil
}

// Send 发送微信通知
func (n *WechatNotifier) Send(ctx context.Context, notification *core.Notification) error {
	// 先验证
	if err := n.Validate(notification); err != nil {
		return fmt.Errorf("wechat validation failed: %w", err)
	}

	// 模拟发送微信消息的过程
	// 在实际项目中，这里会调用微信 API
	appIDPreview := n.appID
	if len(appIDPreview) > 8 {
		appIDPreview = appIDPreview[:8]
	}
	fmt.Printf("[WechatNotifier] Sending WeChat message via API (appID: %s...)\n", appIDPreview)
	fmt.Printf("  OpenID: %s\n", notification.To)
	if notification.Title != "" {
		fmt.Printf("  Title: %s\n", notification.Title)
	}
	fmt.Printf("  Content: %s\n", notification.Content)

	// 模拟网络延迟
	select {
	case <-time.After(80 * time.Millisecond):
		// 模拟发送成功
		fmt.Printf("[WechatNotifier] WeChat message sent successfully\n")
		return nil
	case <-ctx.Done():
		return fmt.Errorf("wechat send cancelled: %w", ctx.Err())
	}
}
