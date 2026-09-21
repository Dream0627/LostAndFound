// 本文件实现错误的“抛出”与“统一处理”。
// 核心思路：handler 只调用 AbortWith... 把错误挂到 gin.Context 上并中断后续处理，
// 真正的响应写出交给全局中间件 ErrorHandler 调用 Handle 完成，做到响应格式统一。
package apperror

import (
	"errors"

	"github.com/gin-gonic/gin"

	"LAF/pkg/response"
)

// InternalErrorKey 是存放到 gin.Context 里的键名，用来保留“原始(内部)错误对象”，
// 便于日志记录与排查；前端看不到它，只能看到对外返回的 Code/Msg。
const InternalErrorKey = "internal_error"

// AbortWithException 用于“业务错误 + 可选原始错误”的场景。
// apiError 是给用户看的业务错误；err 是可选的底层原始错误(如数据库报错)，有则暂存供日志使用。
// 它会写入错误并 Abort 终止后续 handler。
// 特别注意：调用本函数后，调用方必须紧跟 return，
// 否则后续代码仍会继续执行（本项目历史上曾因漏写 return 造成权限绕过）。
func AbortWithException(c *gin.Context, apiError *Error, err error) {
	if err != nil {
		c.Set(InternalErrorKey, err)
	}
	_ = c.Error(apiError)
	c.Abort()
}

// Handle 由全局中间件调用，负责把 error 转换成最终响应。
// 它先假设是服务器错误，再用 errors.As 尝试把它识别为 *Error（业务错误）；
// 识别成功就用业务错误的码与提示，识别失败则统一按“系统异常”返回。
func Handle(c *gin.Context, err error) {
	apiError := ServerError
	var target *Error
	if errors.As(err, &target) {
		apiError = target
	}
	response.Error(c, apiError.Code, apiError.Msg)
}

// AbortWithError 用于传入的是 error 类型（它本身可能就是一个 *Error）的场景，
// 同样先暂存原始错误、写入错误列表并中断；最终的码/提示会在 Handle 里解析。
func AbortWithError(c *gin.Context, apiError error) {
	c.Set(InternalErrorKey, apiError)
	_ = c.Error(apiError)
	c.Abort()
}