// Package user 是用户模块的 HTTP 处理器层：注册、登录、资料、改密。
// 本文件对应“登录”接口。
package user

import (
	"github.com/gin-gonic/gin"

	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// LoginRequest 是登录请求体：用户名与密码，二者必传。
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 是“登录”的处理器工厂。
func Login(userService *service.UserService) gin.HandlerFunc { 
	return func(c *gin.Context) {
		var request LoginRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		result, err := userService.Login(request.Username, request.Password) // 校验账号密码并签发令牌
		if err != nil {
			apperror.AbortWithError(c, err)
			return 
		}

		response.Success(c, result) // 返回令牌与用户信息
	}
}