package main

import (
	"net/http"

	"github.com/stones-hub/taurus-pro-http/pkg/router"
	"{{.ProjectName}}/app"
)

func main() {
	httpServer := app.GetHttpServer()
	httpServer.AddRouter(router.Router{
		Path:    "/home",
		Handler: http.HandlerFunc(app.T.IndexController.Home),
	})
	app.StartAndWait(httpServer)
}
