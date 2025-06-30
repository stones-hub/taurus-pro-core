package http

import (
	"time"

	"github.com/stones-hub/taurus-pro-config/pkg/config"
	"github.com/stones-hub/taurus-pro-core/pkg/components/types"
	"github.com/stones-hub/taurus-pro-http/pkg/server"
)

func ProvideHttpComponent(cfg *config.Config) (*server.Server, error) {
	httpServer := server.NewServer(
		server.WithAddr(cfg.GetString("http.address")+":"+cfg.GetString("http.port")),
		server.WithReadTimeout(time.Duration(cfg.GetInt("http.read_timeout"))*time.Second),
		server.WithWriteTimeout(time.Duration(cfg.GetInt("http.write_timeout"))*time.Second),
		server.WithIdleTimeout(time.Duration(cfg.GetInt("http.idle_timeout"))*time.Second),
	)
	return httpServer, nil
}

var HttpComponent = types.Component{
	Name:         "http",
	Package:      "github.com/stones-hub/taurus-pro-http",
	Version:      "v0.0.2",
	Description:  "HTTP服务器组件",
	IsCustom:     true,
	Required:     true,
	Dependencies: []string{"config"},
	Wire: &types.Wire{
		RequirePath:  "github.com/stones-hub/taurus-pro-http/pkg/server",
		Name:         "Http",
		Type:         "*server.Server",
		ProviderName: "ProvideHttpComponent",
		Provider: `func {{.ProviderName}}(cfg *config.Config) ({{.Type}}, error) {
httpServer := server.NewServer(
	server.WithAddr(cfg.GetString("http.address")+":"+cfg.GetString("http.port")),
	server.WithReadTimeout(time.Duration(cfg.GetInt("http.read_timeout"))*time.Second),
	server.WithWriteTimeout(time.Duration(cfg.GetInt("http.write_timeout"))*time.Second),
	server.WithIdleTimeout(time.Duration(cfg.GetInt("http.idle_timeout"))*time.Second),
)
return httpServer, nil
}`,
	},
}
