package controller

import (
	"net/http"

	"{{.ProjectName}}/app/service"

	"github.com/google/wire"
	"github.com/stones-hub/taurus-pro-http/pkg/httpx"
)

// IndexController 首页控制器
type IndexController struct {
	IndexService *service.IndexService
}

// IndexControllerSet wire provider set
var IndexControllerSet = wire.NewSet(wire.Struct(new(IndexController), "*"))

// Home 处理首页请求
func (c *IndexController) Home(w http.ResponseWriter, r *http.Request) {
	content := c.IndexService.Home()
	httpx.SendResponse(w, http.StatusOK, content, nil)
}
