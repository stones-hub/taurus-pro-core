package controller

import (
	"github.com/stones-hub/taurus-pro-http/server"
)

// ExampleController 示例控制器
type ExampleController struct{}

// NewExampleController 创建示例控制器实例
func NewExampleController() *ExampleController {
	return &ExampleController{}
}

// Register 注册路由
func (c *ExampleController) Register(srv *server.Server) {
	srv.GET("/example", c.HandleExample)
}

// HandleExample 处理示例请求
func (c *ExampleController) HandleExample(ctx *server.Context) {
	ctx.JSON(200, map[string]string{
		"message": "这是一个示例响应",
	})
}
