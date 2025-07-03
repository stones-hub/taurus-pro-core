package consul

import "github.com/stones-hub/taurus-pro-core/pkg/components/types"

var ConsulComponent = types.Component{
	Name:         "consul",
	Package:      "github.com/stones-hub/taurus-pro-consul",
	Version:      "v0.0.1",
	Description:  "consul组件",
	IsCustom:     true,
	Required:     false,
	Dependencies: []string{"config"},
	Wire:         []*types.Wire{consulWire},
}

var consulWire = &types.Wire{
	RequirePath:  []string{"github.com/stones-hub/taurus-pro-consul/pkg/consul"},
	Name:         "Consul",
	Type:         "*consul.ConsulProvider",
	ProviderName: "ProvideConsulComponent",
	Provider:     ``,
}
