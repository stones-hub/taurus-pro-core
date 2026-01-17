package observers

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/stones-hub/taurus-pro-core/example/observer/core"
	"github.com/stones-hub/taurus-pro-core/example/observer/events"
)

// LoggerObserver 日志观察者
// 负责记录所有事件到日志
type LoggerObserver struct {
	id      string
	enabled bool
}

// NewLoggerObserver 创建日志观察者
func NewLoggerObserver(id string) *LoggerObserver {
	return &LoggerObserver{
		id:      id,
		enabled: true,
	}
}

// ID 返回观察者ID
func (o *LoggerObserver) ID() string {
	return o.id
}

// Handle 处理事件
func (o *LoggerObserver) Handle(ctx context.Context, event core.Event) error {
	if !o.enabled {
		return nil
	}

	// 模拟一些处理时间
	time.Sleep(10 * time.Millisecond)

	switch e := event.(type) {
	case *events.UserLoginEvent:
		log.Printf("[Logger] 用户登录: UserID=%s, IP=%s, Time=%s",
			e.UserID, e.IP, e.LoginTime.Format(time.RFC3339))
	case *events.UserLogoutEvent:
		log.Printf("[Logger] 用户登出: UserID=%s, SessionID=%s, Time=%s",
			e.UserID, e.SessionID, e.LogoutTime.Format(time.RFC3339))
	case *events.UserRegisterEvent:
		log.Printf("[Logger] 用户注册: UserID=%s, Username=%s, Email=%s, Time=%s",
			e.UserID, e.Username, e.Email, e.RegisterTime.Format(time.RFC3339))
	case *events.UserUpdateEvent:
		log.Printf("[Logger] 用户更新: UserID=%s, UpdatedFields=%v, Time=%s",
			e.UserID, e.UpdatedFields, e.UpdateTime.Format(time.RFC3339))
	case *events.ConfigChangeEvent:
		log.Printf("[Logger] 配置变更: Key=%s, OldValue=%v, NewValue=%v, Time=%s",
			e.Key, e.OldValue, e.NewValue, e.ChangeTime.Format(time.RFC3339))
	default:
		log.Printf("[Logger] 未知事件: Type=%s, Time=%s",
			event.Type(), event.Timestamp().Format(time.RFC3339))
	}

	return nil
}

// Enable 启用观察者
func (o *LoggerObserver) Enable() {
	o.enabled = true
}

// Disable 禁用观察者
func (o *LoggerObserver) Disable() {
	o.enabled = false
}

// AuditObserver 审计观察者
// 负责将重要事件记录到审计日志
type AuditObserver struct {
	id      string
	auditDB []string // 模拟审计数据库
}

// NewAuditObserver 创建审计观察者
func NewAuditObserver(id string) *AuditObserver {
	return &AuditObserver{
		id:      id,
		auditDB: make([]string, 0),
	}
}

// ID 返回观察者ID
func (o *AuditObserver) ID() string {
	return o.id
}

// Handle 处理事件
func (o *AuditObserver) Handle(ctx context.Context, event core.Event) error {
	// 只审计重要事件
	switch event.Type() {
	case events.EventTypeUserLogin, events.EventTypeUserLogout, events.EventTypeUserRegister:
		auditRecord := fmt.Sprintf("[AUDIT] %s: %s at %s",
			event.Type(), event.Timestamp().Format(time.RFC3339), time.Now().Format(time.RFC3339))
		o.auditDB = append(o.auditDB, auditRecord)
		log.Printf(auditRecord)
	}
	return nil
}

// GetAuditRecords 获取审计记录
func (o *AuditObserver) GetAuditRecords() []string {
	return o.auditDB
}
