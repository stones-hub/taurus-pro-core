package core

// ============================================================================
// 核心层 (Core Layer)
// ============================================================================
// 定义策略模式的接口和注册表，这是整个策略模式的基础
// ============================================================================

import (
	"context"
	"fmt"
)

// Notification 通知消息结构
type Notification struct {
	Title   string            // 标题
	Content string            // 内容
	To      string            // 接收者
	Extras  map[string]string // 扩展信息
}

// Notifier 通知策略接口
// 这是策略模式的核心：定义统一的策略行为
type Notifier interface {
	// Name 返回策略名称，用于注册和识别
	Name() string

	// Send 发送通知的核心方法
	// ctx: 上下文，可用于传递请求信息、超时控制等
	// notification: 通知内容
	// 返回错误信息，nil 表示成功
	Send(ctx context.Context, notification *Notification) error

	// Validate 验证通知参数是否有效
	// 在发送前进行校验，避免无效请求
	Validate(notification *Notification) error
}

// NotifierRegistry 策略注册表
// 使用 map 存储所有注册的策略，key 为策略名称
type NotifierRegistry struct {
	strategies map[string]Notifier
}

// NewNotifierRegistry 创建新的策略注册表
func NewNotifierRegistry() *NotifierRegistry {
	return &NotifierRegistry{
		strategies: make(map[string]Notifier),
	}
}

// Register 注册一个策略
func (r *NotifierRegistry) Register(notifier Notifier) {
	if notifier == nil {
		panic("cannot register nil notifier")
	}
	name := notifier.Name()
	if name == "" {
		panic("notifier name cannot be empty")
	}
	r.strategies[name] = notifier
}

// Get 根据名称获取策略
func (r *NotifierRegistry) Get(name string) (Notifier, error) {
	notifier, ok := r.strategies[name]
	if !ok {
		return nil, fmt.Errorf("notifier strategy '%s' not found", name)
	}
	return notifier, nil
}

// List 列出所有已注册的策略名称
func (r *NotifierRegistry) List() []string {
	names := make([]string, 0, len(r.strategies))
	for name := range r.strategies {
		names = append(names, name)
	}
	return names
}

// Exists 检查策略是否存在
func (r *NotifierRegistry) Exists(name string) bool {
	_, ok := r.strategies[name]
	return ok
}

// 全局注册表实例（可选，也可以使用依赖注入）
var defaultRegistry = NewNotifierRegistry()

// RegisterDefault 注册到全局注册表
func RegisterDefault(notifier Notifier) {
	defaultRegistry.Register(notifier)
}

// GetDefault 从全局注册表获取策略
func GetDefault(name string) (Notifier, error) {
	return defaultRegistry.Get(name)
}

// ListDefault 列出全局注册表的所有策略
func ListDefault() []string {
	return defaultRegistry.List()
}
