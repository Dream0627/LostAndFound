// 本文件对应“查询个人资料”接口(需要登录)。
package user

import (
	"LAF/internal/service"
	"LAF/pkg/response"
	//"errors"
	//"strconv"

	//"fmt"

	//"strings"
	//"time"
	//"LAF/internal/model"
	//"LAF/internal/repository"
	"LAF/internal/middleware"
	"LAF/pkg/apperror"

	"github.com/gin-gonic/gin"
)

// GetProfile 是“查询个人资料”的处理器工厂。
func GetProfile(userService *service.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		nowUserID, _ := middleware.CurrentUserID(c)
		
		user, err := userService.GetProfile(nowUserID) // 用令牌里的用户 ID 查资料，不接受前端指定他人 id
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, user) // 返回用户信息与其帖子
	}
}