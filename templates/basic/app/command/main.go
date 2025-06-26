package main

import (
	"log"

	"github.com/stones-hub/taurus-pro-http/server"
	"{{.PackageName}}/app/controller"
	"{{.PackageName}}/app/middleware"
)

func main() {
	// 创建 HTTP 服务器实例
	srv := server.New()

	// 添加全局中间件
	srv.Use(middleware.Logger())

	// 注册路由
	controller.NewExampleController().Register(srv)

	// 启动服务器
	if err := srv.Start(); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
