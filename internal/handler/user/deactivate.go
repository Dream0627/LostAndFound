// 本文件对应“注销本人账号”接口(需要登录)。
package user

import (
	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// DeactivateAccount 是“注销本人账号”的处理器工厂。
// 通过令牌拿到当前用户 ID，注销只作用于本人，不接受前端指定他人 ID。
func DeactivateAccount(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		nowUserID, _ := middleware.CurrentUserID(c)

		if err := userService.DeactivateAccount(nowUserID); err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, gin.H{"user_id": nowUserID}) // 返回被注销的用户 ID
	}
}
