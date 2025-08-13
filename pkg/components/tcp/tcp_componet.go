package tcp

import (
	"log"
	"time"

	"github.com/stones-hub/taurus-pro-config/pkg/config"
	"github.com/stones-hub/taurus-pro-core/pkg/components/types"
	TCPServer "github.com/stones-hub/taurus-pro-tcp/pkg/tcp"
	"github.com/stones-hub/taurus-pro-tcp/pkg/tcp/protocol"
)

var TcpComponent = types.Component{
	Name:         "tcp",
	Package:      "github.com/stones-hub/taurus-pro-tcp",
	Version:      "v0.0.2",
	Description:  "TCP服务器组件",
	IsCustom:     true,
	Required:     false,
	Dependencies: []string{"config"},
	Wire:         []*types.Wire{tcpWire},
}

var tcpWire = &types.Wire{
	RequirePath:  []string{"TCPServer@github.com/stones-hub/taurus-pro-tcp/pkg/tcp", "github.com/stones-hub/taurus-pro-tcp/pkg/tcp/protocol"},
	Name:         "TCPServer",
	Type:         "*TCPServer.Server",
	ProviderName: "ProvideTcpComponent",
	Provider: `func {{.ProviderName}}(cfg *config.Config) ({{.Type}}, func(), error) {
	enable := cfg.GetBool("tcp.enable")
	if !enable {
		return nil, func() {}, nil
	}

	proto, err := protocol.NewProtocol(protocol.WithType(protocol.ProtocolType(cfg.GetString("tcp.protocol"))))
	if err != nil {
		return nil, func() {}, err
	}

	server, cleanup, err := TCPServer.NewServer(
		cfg.GetString("tcp.address"),
		proto,
		TCPServer.GetHandler(cfg.GetString("tcp.handler")),
		TCPServer.WithMaxConnections(int32(cfg.GetInt("tcp.max_connections"))),
		TCPServer.WithConnectionMaxMessageSize(uint32(cfg.GetInt("tcp.max_message_size"))),
		TCPServer.WithConnectionBufferSize(cfg.GetInt("tcp.buffer_size")),
		TCPServer.WithConnectionIdleTimeout(time.Duration(cfg.GetInt("tcp.idle_timeout"))),
		TCPServer.WithConnectionRateLimiter(cfg.GetInt("tcp.rate_limiter")),
	)

	if err != nil {
		log.Printf("%s🔗 -> Tcp all initialized failed. %s\n", "\033[31m", "\033[0m")
		return nil, func() {}, err
	}

	go func() {
		log.Printf("%s🔗 -> Tcp server start on %s. %s\n", "\033[32m", cfg.GetString("tcp.address"), "\033[0m")
		err := server.Start()
		if err != nil {
			log.Printf("%s🔗 -> Tcp server start failed. %s\n", "\033[31m", "\033[0m")
			return
		}
	}()

	log.Printf("%s🔗 -> Tcp all initialized successfully. %s\n", "\033[32m", "\033[0m")

	return server, func() {
		cleanup()
		log.Printf("%s🔗 -> Clean up tcp components successfully. %s\n", "\033[32m", "\033[0m")
	}, nil
}
`,
}

func ProvideTcpComponent(cfg *config.Config) (*TCPServer.Server, func(), error) {
	enable := cfg.GetBool("tcp.enable")
	if !enable {
		return nil, func() {}, nil
	}

	proto, err := protocol.NewProtocol(protocol.WithType(protocol.ProtocolType(cfg.GetString("tcp.protocol"))))
	if err != nil {
		return nil, func() {}, err
	}

	server, cleanup, err := TCPServer.NewServer(
		cfg.GetString("tcp.address"),
		proto,
		TCPServer.GetHandler(cfg.GetString("tcp.handler")),
		TCPServer.WithMaxConnections(int32(cfg.GetInt("tcp.max_connections"))),
		TCPServer.WithConnectionMaxMessageSize(uint32(cfg.GetInt("tcp.max_message_size"))),
		TCPServer.WithConnectionBufferSize(cfg.GetInt("tcp.buffer_size")),
		TCPServer.WithConnectionIdleTimeout(time.Duration(cfg.GetInt("tcp.idle_timeout"))),
		TCPServer.WithConnectionRateLimiter(cfg.GetInt("tcp.rate_limiter")),
	)

	if err != nil {
		log.Printf("%s🔗 -> Tcp all initialized failed. %s\n", "\033[31m", "\033[0m")
		return nil, func() {}, err
	}

	go func() {
		log.Printf("%s🔗 -> Tcp server start on %s. %s\n", "\033[32m", cfg.GetString("tcp.address"), "\033[0m")
		err := server.Start()
		if err != nil {
			log.Printf("%s🔗 -> Tcp server start failed. %s\n", "\033[31m", "\033[0m")
			return
		}
	}()

	log.Printf("%s🔗 -> Tcp all initialized successfully. %s\n", "\033[32m", "\033[0m")

	return server, func() {
		cleanup()
		log.Printf("%s🔗 -> Clean up tcp components successfully. %s\n", "\033[32m", "\033[0m")
	}, nil
}
