package events

import (
	"time"

	"github.com/stones-hub/taurus-pro-core/example/observer/core"
)

// 用户相关事件类型
const (
	EventTypeUserLogin    core.EventType = "user.login"
	EventTypeUserLogout   core.EventType = "user.logout"
	EventTypeUserRegister core.EventType = "user.register"
	EventTypeUserUpdate   core.EventType = "user.update"
)

// UserLoginEvent 用户登录事件
type UserLoginEvent struct {
	core.BaseEvent
	UserID    string
	LoginTime time.Time
	IP        string
	UserAgent string
}

// NewUserLoginEvent 创建用户登录事件
func NewUserLoginEvent(userID, ip, userAgent string) *UserLoginEvent {
	return &UserLoginEvent{
		BaseEvent: core.NewBaseEvent(EventTypeUserLogin),
		UserID:    userID,
		LoginTime: time.Now(),
		IP:        ip,
		UserAgent: userAgent,
	}
}

// UserLogoutEvent 用户登出事件
type UserLogoutEvent struct {
	core.BaseEvent
	UserID     string
	LogoutTime time.Time
	SessionID  string
}

// NewUserLogoutEvent 创建用户登出事件
func NewUserLogoutEvent(userID, sessionID string) *UserLogoutEvent {
	return &UserLogoutEvent{
		BaseEvent:  core.NewBaseEvent(EventTypeUserLogout),
		UserID:     userID,
		LogoutTime: time.Now(),
		SessionID:  sessionID,
	}
}

// UserRegisterEvent 用户注册事件
type UserRegisterEvent struct {
	core.BaseEvent
	UserID      string
	Username    string
	Email       string
	RegisterTime time.Time
}

// NewUserRegisterEvent 创建用户注册事件
func NewUserRegisterEvent(userID, username, email string) *UserRegisterEvent {
	return &UserRegisterEvent{
		BaseEvent:    core.NewBaseEvent(EventTypeUserRegister),
		UserID:       userID,
		Username:     username,
		Email:        email,
		RegisterTime: time.Now(),
	}
}

// UserUpdateEvent 用户更新事件
type UserUpdateEvent struct {
	core.BaseEvent
	UserID      string
	UpdatedFields map[string]interface{}
	UpdateTime  time.Time
}

// NewUserUpdateEvent 创建用户更新事件
func NewUserUpdateEvent(userID string, updatedFields map[string]interface{}) *UserUpdateEvent {
	return &UserUpdateEvent{
		BaseEvent:      core.NewBaseEvent(EventTypeUserUpdate),
		UserID:         userID,
		UpdatedFields: updatedFields,
		UpdateTime:     time.Now(),
	}
}
