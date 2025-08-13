package config

import "github.com/stones-hub/taurus-pro-core/pkg/components/types"

// 基础组件
var ConfigComponent = types.Component{
	Name:        "config",
	Package:     "github.com/stones-hub/taurus-pro-config",
	Version:     "v0.0.4",
	Description: "配置管理组件",
	IsCustom:    true,
	Required:    true,
	Wire:        []*types.Wire{},
}
