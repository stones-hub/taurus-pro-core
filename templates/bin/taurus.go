package main

import (
	"fmt"
	"log"
	"time"

	"github.com/stones-hub/taurus-pro-config/pkg/config"
	"github.com/stones-hub/taurus-pro-http/pkg/server"
)

func main() {
	// 使用 wire 初始化应用
	application, cleanup, err := app.InitializeApplication()
	if err != nil {
		log.Fatal(err)
	}
	defer cleanup()

	// 通过容器获取组件
	fmt.Println("Application initialized successfully!")

	cfg := config.New()

	cfg.PrintEnable = true

	// 初始化配置
	if err := cfg.Initialize("config/", ".env.local"); err != nil {
		log.Fatalf("Failed to initialize config: %v", err)
	}

	// 注册配置组件
	application.Container.Register("config", cfg)

	// 注册http组件

	srv := server.NewServer(server.WithAddr(cfg.GetString("http.address")+":"+cfg.GetString("http.port")),
		server.WithReadTimeout(time.Duration(cfg.GetInt("http.read_timeout"))*time.Second),
		server.WithWriteTimeout(time.Duration(cfg.GetInt("http.write_timeout"))*time.Second),
		server.WithIdleTimeout(time.Duration(cfg.GetInt("http.idle_timeout"))*time.Second))
	application.Container.Register("http", srv)

	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}
