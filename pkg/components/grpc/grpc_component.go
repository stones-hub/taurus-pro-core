package grpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"google.golang.org/grpc/keepalive"

	"github.com/stones-hub/taurus-pro-config/pkg/config"
	"github.com/stones-hub/taurus-pro-core/pkg/components/types"
	"github.com/stones-hub/taurus-pro-grpc/pkg/grpc/server"
)

func ProvideGrpcComponent(cfg *config.Config) (*server.Server, func(), error) {

	if !cfg.GetBool("grpc.enable") {
		return nil, func() {}, nil
	}

	options := []server.ServerOption{
		server.WithAddress(cfg.GetString("grpc.address")),
		server.WithMaxConns(cfg.GetInt("grpc.max_conns")),
	}

	if cfg.GetBool("grpc.keepalive.enabled") {
		options = append(options, server.WithKeepAlive(&keepalive.ServerParameters{
			MaxConnectionIdle:     time.Duration(cfg.GetInt("grpc.keepalive.max_connection_idle")) * time.Minute,
			MaxConnectionAge:      time.Duration(cfg.GetInt("grpc.keepalive.max_connection_age")) * time.Minute,
			MaxConnectionAgeGrace: time.Duration(cfg.GetInt("grpc.keepalive.max_connection_age_grace")) * time.Second,
			Time:                  time.Duration(cfg.GetInt("grpc.keepalive.time")) * time.Hour,
			Timeout:               time.Duration(cfg.GetInt("grpc.keepalive.timeout")) * time.Second,
		}))
	}

	if cfg.GetBool("grpc.tls.enabled") {

		// 加载服务器证书和私钥
		cert, err := tls.LoadX509KeyPair(cfg.GetString("grpc.tls.crt"), cfg.GetString("grpc.tls.key"))
		if err != nil {
			return nil, func() {}, fmt.Errorf("failed to load key pair: %v", err)
		}

		// 加载 CA 证书用于验证客户端证书
		caCert, err := os.ReadFile(cfg.GetString("grpc.tls.ca"))
		if err != nil {
			return nil, func() {}, fmt.Errorf("failed to read CA certificate: %v", err)
		}

		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caCert) {
			return nil, func() {}, fmt.Errorf("failed to append CA certificate")
		}

		options = append(options, server.WithTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    certPool,
		}))

	}

	return server.NewServer(options...)
}

var grpcWire = &types.Wire{
	RequirePath:  []string{"gRPCServer@github.com/stones-hub/taurus-pro-grpc/pkg/grpc/server", "time", "crypto/tls", "crypto/x509", "os", "fmt", "google.golang.org/grpc/keepalive"},
	Name:         "GRPC",
	Type:         "*gRPCServer.Server",
	ProviderName: "ProvideGrpcComponent",
	Provider: `func {{.ProviderName}}(cfg *config.Config) (*gRPCServer.Server, func(), error) {

	if !cfg.GetBool("grpc.enable") {
		return nil, func() {}, nil
	}

	options := []gRPCServer.ServerOption{
		gRPCServer.WithAddress(cfg.GetString("grpc.address")),
		gRPCServer.WithMaxConns(cfg.GetInt("grpc.max_conns")),
	}

	if cfg.GetBool("grpc.keepalive.enabled") {
		options = append(options, gRPCServer.WithKeepAlive(&keepalive.ServerParameters{
			MaxConnectionIdle:     time.Duration(cfg.GetInt("grpc.keepalive.max_connection_idle")) * time.Minute,
			MaxConnectionAge:      time.Duration(cfg.GetInt("grpc.keepalive.max_connection_age")) * time.Minute,
			MaxConnectionAgeGrace: time.Duration(cfg.GetInt("grpc.keepalive.max_connection_age_grace")) * time.Second,
			Time:                  time.Duration(cfg.GetInt("grpc.keepalive.time")) * time.Hour,
			Timeout:               time.Duration(cfg.GetInt("grpc.keepalive.timeout")) * time.Second,
		}))
	}

	if cfg.GetBool("grpc.tls.enabled") {

		// 加载服务器证书和私钥
		cert, err := tls.LoadX509KeyPair(cfg.GetString("grpc.tls.crt"), cfg.GetString("grpc.tls.key"))
		if err != nil {
			return nil, func() {}, fmt.Errorf("failed to load key pair: %v", err)
		}

		// 加载 CA 证书用于验证客户端证书
		caCert, err := os.ReadFile(cfg.GetString("grpc.tls.ca"))
		if err != nil {
			return nil, func() {}, fmt.Errorf("failed to read CA certificate: %v", err)
		}

		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(caCert) {
			return nil, func() {}, fmt.Errorf("failed to append CA certificate")
		}

		options = append(options, gRPCServer.WithTLS(&tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    certPool,
		}))

	}

	return gRPCServer.NewServer(options...)
}`,
}

var GrpcComponent = types.Component{
	Name:         "grpc",
	Package:      "github.com/stones-hub/taurus-pro-grpc",
	Version:      "v0.0.1",
	Description:  "gRPC服务器组件",
	IsCustom:     true,
	Required:     false,
	Dependencies: []string{"config"},
	Wire:         []*types.Wire{grpcWire},
}
