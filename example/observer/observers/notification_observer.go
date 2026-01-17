package observers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/stones-hub/taurus-pro-core/example/observer/core"
	"github.com/stones-hub/taurus-pro-core/example/observer/events"
)

// NotificationObserver 通知观察者
// 负责发送通知（邮件、短信等）
type NotificationObserver struct {
	id       string
	enabled  bool
	channels []string // 通知渠道：email, sms, wechat
}

// NewNotificationObserver 创建通知观察者
func NewNotificationObserver(id string, channels []string) *NotificationObserver {
	return &NotificationObserver{
		id:       id,
		enabled:  true,
		channels: channels,
	}
}

// ID 返回观察者ID
func (o *NotificationObserver) ID() string {
	return o.id
}

// Handle 处理事件
func (o *NotificationObserver) Handle(ctx context.Context, event core.Event) error {
	if !o.enabled {
		return nil
	}

	// 模拟发送通知的耗时
	time.Sleep(50 * time.Millisecond)

	switch e := event.(type) {
	case *events.UserRegisterEvent:
		// 新用户注册，发送欢迎邮件
		message := fmt.Sprintf("欢迎 %s 注册！您的邮箱是 %s", e.Username, e.Email)
		o.sendNotification("email", e.Email, "欢迎注册", message)
		log.Printf("[Notification] 已发送欢迎邮件给: %s", e.Email)

	case *events.UserLoginEvent:
		// 用户登录，发送安全通知（如果配置了）
		if o.hasChannel("sms") {
			message := fmt.Sprintf("您的账户在 %s 登录，IP: %s", time.Now().Format(time.RFC3339), e.IP)
			o.sendNotification("sms", "", "登录通知", message)
			log.Printf("[Notification] 已发送登录通知短信")
		}
	}

	return nil
}

// sendNotification 发送通知（模拟）
func (o *NotificationObserver) sendNotification(channel, to, subject, message string) {
	// 这里应该是实际的发送逻辑
	log.Printf("[Notification] 通过 %s 发送通知到 %s: %s - %s", channel, to, subject, message)
}

// hasChannel 检查是否配置了指定渠道
func (o *NotificationObserver) hasChannel(channel string) bool {
	for _, ch := range o.channels {
		if ch == channel {
			return true
		}
	}
	return false
}

// Enable 启用观察者
func (o *NotificationObserver) Enable() {
	o.enabled = true
}

// Disable 禁用观察者
func (o *NotificationObserver) Disable() {
	o.enabled = false
}
