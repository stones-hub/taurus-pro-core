package http

import (
	"log"
	"time"

	"github.com/stones-hub/taurus-pro-config/pkg/config"
	"github.com/stones-hub/taurus-pro-core/pkg/components/types"
	"github.com/stones-hub/taurus-pro-http/pkg/mcp"
	"github.com/stones-hub/taurus-pro-http/pkg/server"
	"github.com/stones-hub/taurus-pro-http/pkg/wsocket"
)

func ProvideHttpComponent(cfg *config.Config) (*server.Server, error) {
	httpServer := server.NewServer(
		server.WithAddr(cfg.GetString("http.address")+":"+cfg.GetString("http.port")),
		server.WithReadTimeout(time.Duration(cfg.GetInt("http.read_timeout"))*time.Second),
		server.WithWriteTimeout(time.Duration(cfg.GetInt("http.write_timeout"))*time.Second),
		server.WithIdleTimeout(time.Duration(cfg.GetInt("http.idle_timeout"))*time.Second),
	)

	if cfg.GetBool("websocket.enable") {
		wsocket.Initialize()
		log.Printf("%s🔗 -> http-websocket initialized successfully. %s\n", "\033[32m", "\033[0m")
	}

	log.Printf("%s🔗 -> Http all initialized successfully. %s\n", "\033[32m", "\033[0m")

	return httpServer, nil
}

var httpWire = &types.Wire{
	RequirePath:  []string{"github.com/stones-hub/taurus-pro-http/pkg/server", "log", "time", "github.com/stones-hub/taurus-pro-http/pkg/wsocket"},
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

if cfg.GetBool("websocket.enable") {
		wsocket.Initialize()
		log.Printf("%s🔗 -> http-websocket initialized successfully. %s\n", "\033[32m", "\033[0m")
	}

log.Printf("%s🔗 -> Http all initialized successfully. %s\n", "\033[32m", "\033[0m")

return httpServer, nil
}`,
}

func ProvideMcpComponent(cfg *config.Config, httpServer *server.Server) (*mcp.MCPServer, error) {

	// 如果是stdio模式的mcp，不要在http-server中启用, 因为我没构建的就是一个http服务器集群
	if !cfg.GetBool("mcp.enable") || mcp.Transport(cfg.GetString("mcp.transport")) == mcp.TransportStdio {
		return nil, nil
	}

	mcpServer, cleanup, err := mcp.New(
		mcp.WithName("taurus"),
		mcp.WithVersion("v0.0.1"),
		mcp.WithTransport(mcp.Transport(cfg.GetString("mcp.transport"))),
		mcp.WithMode(mcp.Mode(cfg.GetString("mcp.mode"))),
		mcp.WithHttpServer(httpServer),
	)

	if err != nil {
		return nil, err
	}

	if httpServer != nil {
		httpServer.RegisterOnShutdown(func() {
			log.Printf("%s🔗 -> Http-MCP starting shutdown. %s\n", "\033[32m", "\033[0m")
			cleanup()
		})
	}

	return mcpServer, nil
}

var mcpWire = &types.Wire{
	RequirePath:  []string{"github.com/stones-hub/taurus-pro-http/pkg/mcp", "log"},
	Name:         "McpServer",
	Type:         "*mcp.MCPServer",
	ProviderName: "ProvideMcpComponent",
	Provider: `func {{.ProviderName}}(cfg *config.Config, httpServer *server.Server) ({{.Type}}, error) {

	// 如果是stdio模式的mcp，不要在http-server中启用, 因为我没构建的就是一个http服务器集群
	if !cfg.GetBool("mcp.enable") || mcp.Transport(cfg.GetString("mcp.transport")) == mcp.TransportStdio {
		return nil, nil
	}

	mcpServer, cleanup, err := mcp.New(
		mcp.WithName("taurus"),
		mcp.WithVersion("v0.0.1"),
		mcp.WithTransport(mcp.Transport(cfg.GetString("mcp.transport"))),
		mcp.WithMode(mcp.Mode(cfg.GetString("mcp.mode"))),
		mcp.WithHttpServer(httpServer),
	)

	if err != nil {
		return nil, err
	}

	if httpServer != nil {
		httpServer.RegisterOnShutdown(func() {
			log.Printf("%s🔗 -> Http-MCP starting shutdown. %s\n", "\033[32m", "\033[0m")
			cleanup()
		})
	}

	return mcpServer, nil
}`,
}

var HttpComponent = types.Component{
	Name:         "http",
	Package:      "github.com/stones-hub/taurus-pro-http",
	Version:      "v0.0.7",
	Description:  "Http,WebSocket服务器组件",
	IsCustom:     true,
	Required:     true,
	Dependencies: []string{"config"},
	Wire:         []*types.Wire{httpWire, mcpWire},
}
