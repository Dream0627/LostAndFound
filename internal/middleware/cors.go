// 本文件实现跨域(CORS)中间件，用于前后端分离部署场景。
// 单端口同源部署(前端 Nginx 反代后端)时浏览器不触发跨域，可不配置任何来源；
// 前后端分端口部署(如前端 5173、后端 8080)时，必须在配置里列出前端来源，否则浏览器会拦截请求。
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS 返回按白名单校验来源的跨域中间件。
// allowOrigins 来自配置 server.cors_allow_origins；为空表示不放开任何跨域来源。
// 对不在白名单中的 Origin：不回跨域响应头，浏览器端自然会拦截，服务端逻辑不受影响。
func CORS(allowOrigins []string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(allowOrigins))
	for _, origin := range allowOrigins {
		origin = strings.TrimSpace(origin)
		if origin != "" {
			allowed[origin] = struct{}{}
		}
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")
		// 非跨域请求(无 Origin 头)或未配置白名单时直接放行，不加任何跨域响应头。
		if origin != "" && len(allowed) > 0 {
			if _, ok := allowed[origin]; ok {
				// 回具体来源而非 *：配合 JWT 的 Authorization 头使用更安全，也兼容凭据场景。
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
				c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")
			}
		}

		// 预检请求(OPTIONS)到此为止，不进入业务路由。
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}
