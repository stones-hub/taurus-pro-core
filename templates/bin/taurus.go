//lint:file-ignore ST1001 dot imports are allowed here

package main

import (
	"fmt"
	"net/http"

	"{{.ProjectName}}/app"
	"{{.ProjectName}}/internal/taurus"

	"github.com/stones-hub/taurus-pro-http/pkg/middleware"
	"github.com/stones-hub/taurus-pro-http/pkg/router"
)

func main() {
	pprof()

	taurus.Container.Http.AddRouter(router.Router{
		Path:    "/home",
		Handler: http.HandlerFunc(app.Core.IndexController.Home),
		Middleware: []router.MiddlewareFunc{
			middleware.RecoveryMiddleware(func(err any, stack string) {
				fmt.Printf("Error: %v\nStack: %s\n", err, stack)
			}),
		},
	})

	taurus.Container.Http.AddRouter(router.Router{
		Path: "/health",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("health"))
		}),
	})

	taurus.Container.Http.AddRouter(router.Router{
		Path: "/health1",
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("health1"))
		}),
	})

	app.Run()
}

// pprof 路由, 用于测试内存泄漏
func pprof() {
	// 添加内存测试路由
	taurus.Container.Http.AddRouter(router.Router{
		Path:    "/memory/allocate",
		Handler: http.HandlerFunc(app.Core.MemoryController.AllocateMemory),
		Middleware: []router.MiddlewareFunc{
			middleware.RecoveryMiddleware(func(err any, stack string) {
				fmt.Printf("Error: %v\nStack: %s\n", err, stack)
			}),
		},
	})

	taurus.Container.Http.AddRouter(router.Router{
		Path:    "/memory/leak",
		Handler: http.HandlerFunc(app.Core.MemoryController.SimulateMemoryLeak),
		Middleware: []router.MiddlewareFunc{
			middleware.RecoveryMiddleware(func(err any, stack string) {
				fmt.Printf("Error: %v\nStack: %s\n", err, stack)
			}),
		},
	})

	taurus.Container.Http.AddRouter(router.Router{
		Path:    "/memory/free",
		Handler: http.HandlerFunc(app.Core.MemoryController.FreeMemory),
		Middleware: []router.MiddlewareFunc{
			middleware.RecoveryMiddleware(func(err any, stack string) {
				fmt.Printf("Error: %v\nStack: %s\n", err, stack)
			}),
		},
	})
}
