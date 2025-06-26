package container

import (
	"fmt"
	"sync"

	"github.com/google/wire"
)

// Container 是应用程序的组件容器
type Container struct {
	mu         sync.RWMutex
	components map[string]interface{}
}

// NewContainer 创建一个新的容器实例
func NewContainer() *Container {
	return &Container{
		components: make(map[string]interface{}),
	}
}

// Register 注册一个组件到容器
func (c *Container) Register(name string, component interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.components[name] = component
}

// Get 从容器中获取一个组件
func (c *Container) Get(name string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	component, ok := c.components[name]
	return component, ok
}

// MustGet 从容器中获取一个组件，如果不存在则panic
func (c *Container) MustGet(name string) interface{} {
	if component, ok := c.Get(name); ok {
		return component
	}
	panic(fmt.Sprintf("component %s not found", name))
}

var (
	// ProviderSet 是容器的provider set
	ContainerSet = wire.NewSet(NewContainer)
)
