package main

import (
	"net/http"

	"{{.ProjectName}}/app"

	"github.com/stones-hub/taurus-pro-http/pkg/router"
)

func main() {

	app.T.Http.AddRouter(router.Router{
		Path:    "/home",
		Handler: http.HandlerFunc(app.T.IndexController.Home),
	})
	app.StartAndWait(app.T.Http)
}
