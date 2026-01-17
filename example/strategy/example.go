package strategy

import (
	"context"
	"fmt"
	"time"

	"github.com/stones-hub/taurus-pro-core/example/strategy/core"
	"github.com/stones-hub/taurus-pro-core/example/strategy/service"
	"github.com/stones-hub/taurus-pro-core/example/strategy/strategies"
)

// ExampleUsage 演示策略模式的使用
func ExampleUsage() {
	fmt.Println("=== 策略模式示例：通知发送系统 ===\n")

	// 1. 创建策略注册表
	registry := core.NewNotifierRegistry()

	// 2. 创建并注册多个策略
	emailNotifier := strategies.NewEmailNotifier("smtp.example.com", 587)
	smsNotifier := strategies.NewSMSNotifier("api_key_123456", "api_secret_789")
	wechatNotifier := strategies.NewWechatNotifier("wx_app_id_123", "wx_app_secret_456")

	registry.Register(emailNotifier)
	registry.Register(smsNotifier)
	registry.Register(wechatNotifier)

	fmt.Printf("已注册的策略: %v\n\n", registry.List())

	// 3. 创建通知服务（依赖注册表，不依赖具体策略）
	svc := service.NewNotificationService(registry)

	// 4. 使用不同的策略发送通知
	ctx := context.Background()

	// 示例 1: 使用邮件策略
	fmt.Println("--- 示例 1: 发送邮件通知 ---")
	emailNotification := &core.Notification{
		Title:   "欢迎注册",
		Content: "感谢您注册我们的服务！",
		To:      "user@example.com",
	}
	if err := svc.SendNotification(ctx, "email", emailNotification); err != nil {
		fmt.Printf("发送失败: %v\n", err)
	}
	fmt.Println()

	// 示例 2: 使用短信策略
	fmt.Println("--- 示例 2: 发送短信通知 ---")
	smsNotification := &core.Notification{
		Content: "您的验证码是：123456，5分钟内有效",
		To:      "13800138000",
	}
	if err := svc.SendNotification(ctx, "sms", smsNotification); err != nil {
		fmt.Printf("发送失败: %v\n", err)
	}
	fmt.Println()

	// 示例 3: 使用微信策略
	fmt.Println("--- 示例 3: 发送微信通知 ---")
	wechatNotification := &core.Notification{
		Title:   "订单通知",
		Content: "您的订单已发货，请注意查收",
		To:      "wx_openid_12345678901234567890",
	}
	if err := svc.SendNotification(ctx, "wechat", wechatNotification); err != nil {
		fmt.Printf("发送失败: %v\n", err)
	}
	fmt.Println()

	// 示例 4: 批量发送（使用多个策略）
	fmt.Println("--- 示例 4: 批量发送通知（多策略） ---")
	batchNotification := &core.Notification{
		Title:   "系统维护通知",
		Content: "系统将于今晚 22:00-24:00 进行维护",
		To:      "user@example.com",
		Extras: map[string]string{
			"phone":  "13800138000",
			"openid": "wx_openid_12345678901234567890",
		},
	}
	results := svc.SendNotificationBatch(ctx, []string{"email", "sms", "wechat"}, batchNotification)
	for strategy, err := range results {
		if err != nil {
			fmt.Printf("策略 '%s' 发送失败: %v\n", strategy, err)
		} else {
			fmt.Printf("策略 '%s' 发送成功\n", strategy)
		}
	}
	fmt.Println()

	// 示例 5: 带降级的发送
	fmt.Println("--- 示例 5: 带降级策略的发送 ---")
	fallbackNotification := &core.Notification{
		Title:   "重要通知",
		Content: "这是一条重要消息",
		To:      "user@example.com",
	}
	// 如果邮件发送失败，自动使用短信
	if err := svc.SendNotificationWithFallback(ctx, "email", "sms", fallbackNotification); err != nil {
		fmt.Printf("所有策略都失败: %v\n", err)
	}
	fmt.Println()

	// 示例 6: 验证通知参数
	fmt.Println("--- 示例 6: 验证通知参数 ---")
	invalidNotification := &core.Notification{
		Content: "测试",
		To:      "", // 缺少接收者
	}
	if err := svc.ValidateNotification("email", invalidNotification); err != nil {
		fmt.Printf("验证失败: %v\n", err)
	}
	fmt.Println()

	// 示例 7: 使用全局注册表
	fmt.Println("--- 示例 7: 使用全局注册表 ---")
	core.RegisterDefault(emailNotifier)
	core.RegisterDefault(smsNotifier)
	notifier, err := core.GetDefault("email")
	if err == nil {
		notification := &core.Notification{
			Title:   "全局注册表示例",
			Content: "使用全局注册表发送",
			To:      "user@example.com",
		}
		notifier.Send(ctx, notification)
	}
	fmt.Println()
}

// ExampleWithContext 演示如何使用 context 控制策略执行
func ExampleWithContext() {
	fmt.Println("=== 使用 Context 控制策略执行 ===\n")

	registry := core.NewNotifierRegistry()
	registry.Register(strategies.NewEmailNotifier("smtp.example.com", 587))
	svc := service.NewNotificationService(registry)

	// 创建一个带超时的 context
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	notification := &core.Notification{
		Title:   "超时测试",
		Content: "这条消息可能会因为超时而取消",
		To:      "user@example.com",
	}

	// 由于超时时间（50ms）小于模拟的发送时间（100ms），应该会超时
	err := svc.SendNotification(ctx, "email", notification)
	if err != nil {
		fmt.Printf("预期中的超时错误: %v\n", err)
	}
	fmt.Println()
}
