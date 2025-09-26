package storage

import "github.com/stones-hub/taurus-pro-core/pkg/components/types"

var StorageComponent = types.Component{
	Name:         "storage",
	Package:      "github.com/stones-hub/taurus-pro-storage",
	Version:      " v0.1.33",
	Description:  "DB、Redis存储, 异步队列组件",
	IsCustom:     true,
	Required:     false,
	Dependencies: []string{"config", "common"},
	Wire:         []*types.Wire{dbWire, redisWire},
}
