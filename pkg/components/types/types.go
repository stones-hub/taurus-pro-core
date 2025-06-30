package types

type Wire struct {
	RequirePath  string // 依赖的包路径
	Name         string // wire中初始化组件的名称
	Type         string // 组件类型
	ProviderName string // 提供者名称
	Provider     string // 提供者函数，如 func ProvideHttpComponent(cfg *config.Config) (*server.Server, error)
}

// Component 表示一个组件
type Component struct {
	Name         string   // 组件别名，如 "config"
	Package      string   // 组件包名，如 "github.com/stones-hub/taurus-pro-config"
	Version      string   // 组件版本，如 "v0.0.1"
	Description  string   // 组件描述
	Required     bool     // 是否为必需组件
	Dependencies []string // 依赖的其他组件别名
	IsCustom     bool     // 是否为自定义组件
	Wire         *Wire
}
