// 本文件实现全局错误处理中间件。
// 它注册在最外层(engine.Use)，是所有请求处理链的“兜底”：
// 让各 handler 只需把错误写进 Context，最终由这里统一转成响应，保证失败响应格式一致。
package middleware

import (
	"github.com/gin-gonic/gin"

	"LAF/pkg/apperror"
)

// ErrorHandler 先放行后续处理(c.Next())，等整条链结束后再检查 Context 上累积的错误：
// 没有错误说明处理正常(响应已由 handler 写出)，直接返回；
// 有错误则交给 apperror.Handle 解析成业务码与提示后写出统一信封。
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 { // 没有错误说明处理正常，无需额外处理
			return
		}

		apperror.Handle(c, c.Errors.Last().Err)
	}
}