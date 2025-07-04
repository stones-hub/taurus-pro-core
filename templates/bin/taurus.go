//lint:file-ignore ST1001 dot imports are allowed here

package main

import (
	"net/http"

	"{{.ProjectName}}/app"
	. "{{.ProjectName}}/app/constants"

	"github.com/stones-hub/taurus-pro-http/pkg/router"
)

func main() {

	Taurus.Http.AddRouter(router.Router{
		Path:    "/home",
		Handler: http.HandlerFunc(app.T.IndexController.Home),
	})

	Taurus.Http.AddRouter(router.Router{
		Path: "/health",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("health"))
		}),
	})

	Taurus.Http.AddRouter(router.Router{
		Path: "/health1",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("health1"))
		}),
	})

	app.Run()
}
