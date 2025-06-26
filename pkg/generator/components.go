package generator

import (
	"fmt"
	"strings"
)

// Component 表示一个组件
type Component struct {
	Name         string   // 组件别名，如 "config"
	Package      string   // 组件包名，如 "github.com/stones-hub/taurus-pro-config"
	Version      string   // 组件版本，如 "v0.0.1"
	Description  string   // 组件描述
	Required     bool     // 是否为必需组件
	Dependencies []string // 依赖的其他组件别名
	IsCustom     bool     // 是否为自定义组件
}

// 所有可用的组件定义
var (
	// 基础组件
	configComponent = Component{
		Name:        "config",
		Package:     "github.com/stones-hub/taurus-pro-config",
		Version:     "v0.0.1",
		Description: "配置管理组件",
		IsCustom:    true,
		Required:    true,
	}

	httpComponent = Component{
		Name:         "http",
		Package:      "github.com/stones-hub/taurus-pro-http",
		Version:      "v0.0.1",
		Description:  "HTTP服务器组件",
		IsCustom:     true,
		Required:     true,
		Dependencies: []string{"config"},
	}

	wireComponent = Component{
		Name:        "wire",
		Package:     "github.com/google/wire",
		Version:     "v0.5.0",
		Description: "依赖注入工具",
		IsCustom:    false,
		Required:    true,
	}

	redisComponent = Component{
		Name:        "redis",
		Package:     "github.com/go-redis/redis/v8",
		Version:     "v8.11.5",
		Description: "Redis客户端组件",
		IsCustom:    false,
		Required:    false,
	}

	// 所有组件列表
	AllComponents = []Component{
		configComponent,
		httpComponent,
		wireComponent,
		redisComponent,
	}
)

// GetRequiredComponents 获取所有必需组件
func GetRequiredComponents() []Component {
	var required []Component
	for _, comp := range AllComponents {
		if comp.Required {
			required = append(required, comp)
		}
	}
	return required
}

// GetOptionalComponents 获取所有可选组件
func GetOptionalComponents() []Component {
	var optional []Component
	for _, comp := range AllComponents {
		if !comp.Required {
			optional = append(optional, comp)
		}
	}
	return optional
}

// GetComponentByName 根据组件名获取组件
func GetComponentByName(name string) (Component, bool) {
	for _, comp := range AllComponents {
		if comp.Name == name {
			return comp, true
		}
	}
	return Component{}, false
}

// ValidateComponents 验证组件依赖关系
func ValidateComponents(selectedComponents []string) error {
	selected := make(map[string]bool)
	for _, name := range selectedComponents {
		selected[name] = true
	}

	// 检查每个选择的组件的依赖是否满足
	for _, name := range selectedComponents {
		for _, comp := range AllComponents {
			if comp.Name == name {
				for _, dep := range comp.Dependencies {
					if !selected[dep] {
						return fmt.Errorf("组件 %s 依赖 %s，但未选择", name, dep)
					}
				}
				break
			}
		}
	}

	return nil
}

// GenerateGoModRequires 生成go.mod的require部分
func GenerateGoModRequires(selectedComponents []string) string {
	requires := []string{
		"require (",
	}

	// 添加基础组件
	for _, comp := range AllComponents {
		if comp.Required {
			requires = append(requires, "\t"+comp.Package+" "+comp.Version)
		}
	}

	requires = append(requires, ")")
	return strings.Join(requires, "\n")
}
