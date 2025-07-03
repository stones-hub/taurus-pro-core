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

	app.T.Http.AddRouter(router.Router{
		Path: "/health",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("health"))
		}),
	})

	app.T.Http.AddRouter(router.Router{
		Path: "/health1",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("health1"))
		}),
	})

	app.StartAndWait(app.T.Http)
}
