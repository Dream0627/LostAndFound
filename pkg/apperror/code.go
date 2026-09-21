// Package apperror 集中定义项目的业务错误。
// 每个业务错误都带一个业务码(Code)和一句提示(Msg)。
// 各层代码只需返回这些预定义错误（或包装错误），由全局中间件统一转成 HTTP 响应，
// 从而避免在业务代码里到处手动拼接错误响应，保证风格统一、便于维护。
package apperror

import (
	"fmt"
)

// Error 是本项目统一的业务错误类型，实现了 Go 标准库的 error 接口。
// Code 对应响应体里的业务码，Msg 是给用户看的提示文字。
type Error struct {
	Code int
	Msg  string
}

// Error 实现 error 接口（这样 *Error 就能当作普通 error 使用）。
// 返回的是“码+提示”的描述文本，主要用于日志排查；
// 真正返回给前端的是 Code 与 Msg，而不是这个字符串。
func (e *Error) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Msg)
}

// NewError 是 Error 的构造函数，方便下面用一行声明一个业务错误。
func NewError(code int, msg string) *Error {
	return &Error{
		Code: code,
		Msg:  msg,
	}
}

// 集中声明全部业务错误。取值约定：
// 400 参数/入参问题，401 未登录或凭证无效，403 权限不足，404 资源不存在，
// 409 资源冲突(如用户名已存在)，1000 服务器内部异常。
var (
	ServerError     = NewError(1000, "系统异常")
	ParamError      = NewError(400, "参数校验失败")
	UnauthorizedError = NewError(401, "未登录或令牌无效")
	LoginError      = NewError(401, "账号或密码错误")
	UserForbiddenError   = NewError(403, "无权删除他人的帖子")
	AdminForbiddenError  = NewError(403, "仅管理员可操作")
	NotFoundError   = NewError(404, "帖子不存在")
	UserRepeatError     = NewError(409, "用户名已存在")
	PostRepeatError     = NewError(409, "帖子已存在")
	UserNotFoundError = NewError(404, "用户不存在")
	OldPasswordError  = NewError(400, "原密码错误")
	CommentNotFoundError = NewError(404, "评论不存在")
	InvalidPostStatusError = NewError(400, "无效的帖子状态")
	PostNotPendingError    = NewError(400, "该帖子不可审核")
)