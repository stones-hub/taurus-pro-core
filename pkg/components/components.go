package components

import (
	"fmt"
	"strings"

	"github.com/stones-hub/taurus-pro-core/pkg/components/common"
	"github.com/stones-hub/taurus-pro-core/pkg/components/config"
	"github.com/stones-hub/taurus-pro-core/pkg/components/http"
	"github.com/stones-hub/taurus-pro-core/pkg/components/storage"
	"github.com/stones-hub/taurus-pro-core/pkg/components/types"
)

// 所有可用的组件定义
var (
	wireComponent = types.Component{
		Name:        "wire",
		Package:     "github.com/google/wire",
		Version:     "v0.5.0",
		Description: "依赖注入工具",
		IsCustom:    false,
		Required:    true,
		Wire:        nil,
	}

	// 所有组件列表
	AllComponents = []types.Component{
		config.ConfigComponent,
		http.HttpComponent,
		common.CommonComponent,
		storage.StorageComponent,
		wireComponent,
	}
)

// GetRequiredComponents 获取所有必需组件
func GetRequiredComponents() []types.Component {
	var required []types.Component
	for _, comp := range AllComponents {
		if comp.Required {
			required = append(required, comp)
		}
	}
	return required
}

// GetOptionalComponents 获取所有可选组件
func GetOptionalComponents() []types.Component {
	var optional []types.Component
	for _, comp := range AllComponents {
		if !comp.Required {
			optional = append(optional, comp)
		}
	}
	return optional
}

// GetComponentByName 根据组件名获取组件
func GetComponentByName(name string) (types.Component, bool) {
	for _, comp := range AllComponents {
		if comp.Name == name {
			return comp, true
		}
	}
	return types.Component{}, false
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
