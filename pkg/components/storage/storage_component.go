package storage

import "github.com/stones-hub/taurus-pro-core/pkg/components/types"

var StorageComponent = types.Component{
	Name:         "storage",
	Package:      "github.com/stones-hub/taurus-pro-storage",
	Version:      "v0.0.7",
	Description:  "DB,Redis存储组件",
	IsCustom:     true,
	Required:     false,
	Dependencies: []string{"config"},
	Wire:         []*types.Wire{dbWire, redisWire},
}
