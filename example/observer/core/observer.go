package core

import "context"

// Observer 观察者接口
// 所有观察者必须实现此接口来处理事件
type Observer interface {
	// ID 返回观察者的唯一标识
	ID() string
	// Handle 处理事件，返回错误表示处理失败
	Handle(ctx context.Context, event Event) error
}

// ObserverFunc 函数式观察者，将函数转换为观察者接口
// 这是 Go 风格的实现方式，更灵活易用
type ObserverFunc func(ctx context.Context, event Event) error

// ID 返回函数观察者的标识（使用函数地址作为ID）
func (f ObserverFunc) ID() string {
	return "observer_func"
}

// Handle 执行观察者函数
func (f ObserverFunc) Handle(ctx context.Context, event Event) error {
	return f(ctx, event)
}

// NamedObserverFunc 带名称的函数式观察者
type NamedObserverFunc struct {
	id   string
	fn   func(ctx context.Context, event Event) error
}

// NewNamedObserverFunc 创建带名称的函数式观察者
func NewNamedObserverFunc(id string, fn func(ctx context.Context, event Event) error) *NamedObserverFunc {
	return &NamedObserverFunc{
		id: id,
		fn: fn,
	}
}

// ID 返回观察者ID
func (n *NamedObserverFunc) ID() string {
	return n.id
}

// Handle 处理事件
func (n *NamedObserverFunc) Handle(ctx context.Context, event Event) error {
	return n.fn(ctx, event)
}
