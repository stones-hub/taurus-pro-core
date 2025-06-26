package middleware

import (
	"log"
	"time"

	"github.com/stones-hub/taurus-pro-http/server"
)

// Logger 创建一个日志中间件
func Logger() server.HandlerFunc {
	return func(ctx *server.Context) {
		start := time.Now()

		// 处理请求
		ctx.Next()

		// 记录请求信息
		log.Printf("[%s] %s %s %v",
			ctx.Request.Method,
			ctx.Request.URL.Path,
			ctx.Request.RemoteAddr,
			time.Since(start),
		)
	}
}
