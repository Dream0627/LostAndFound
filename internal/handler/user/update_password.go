// 本文件对应“修改密码”接口(需要登录)。
package user

import (
	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// UpdatePasswordRequest 是改密请求体：原密码、新密码、确认密码，均必传。
type UpdatePasswordRequest struct {
	OldPassword     string `json:"old_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required"`
	ConfirmPassword string `json:"confirm_password" binding:"required"`
}

// UpdatePassword 是“修改密码”的处理器工厂。
func UpdatePassword(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		nowUserID, _ := middleware.CurrentUserID(c)

		var request UpdatePasswordRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		input := service.UpdatePasswordInput{ // 组装业务入参
			OldPassword:     request.OldPassword,
			NewPassword:     request.NewPassword,
			ConfirmPassword: request.ConfirmPassword,
		}
		if err := userService.UpdatePassword(nowUserID, input); err != nil { // 交由业务层校验并更新密码
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, nil) // 改密成功返回空数据
	}
}
