package apperror

import (
	"fmt"
)

type Error struct {
	Code int
	Msg  string
}

func (e *Error) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Msg)
}

func NewError(code int, msg string) *Error {
	return &Error{
		Code: code,
		Msg:  msg,
	}
}

var (
	ServerError     = NewError(1000, "系统异常")
	ParamError      = NewError(400, "参数校验失败")
	UnauthorizedError = NewError(401, "未登录或令牌无效")
	LoginError      = NewError(401, "账号或密码错误")
	UserForbiddenError   = NewError(403, "无权删除他人的帖子")
	AdminForbiddenError  = NewError(403, "仅管理员可删除任意帖子")
	NotFoundError   = NewError(404, "帖子不存在")
	UserRepeatError     = NewError(409, "用户名已存在")
	PostRepeatError     = NewError(409, "帖子已存在")
	UserNotFoundError = NewError(404, "用户不存在")
)