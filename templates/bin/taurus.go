//lint:file-ignore ST1001 dot imports are allowed here

package main

import (
	"fmt"
	"net/http"

	"github.com/stones-hub/taurus-pro-http/pkg/middleware"
	"github.com/stones-hub/taurus-pro-http/pkg/router"
	"{{.ProjectName}}/app"
)

func main() {
	app.T.Http.AddRouter(router.Router{
		Path:    "/home",
		Handler: http.HandlerFunc(app.T.IndexController.Home),
		Middleware: []router.MiddlewareFunc{
			middleware.RecoveryMiddleware(func(err any, stack string) {
				fmt.Printf("Error: %v\nStack: %s\n", err, stack)
			}),
		},
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

	app.Run()
}
