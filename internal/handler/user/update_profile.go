// 本文件对应“更新个人资料”接口(需要登录)。
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

// UpdateProfile 是“更新个人资料”的处理器工厂。
func UpdateProfile(userService *service.UserService) gin.HandlerFunc { 
	return func(c *gin.Context) {
		nowUserID, _ := middleware.CurrentUserID(c)
		
		var input service.UpdateProfileInput // 直接把请求体绑定到业务入参结构(字段名一致)
		if err := c.ShouldBindJSON(&input); err != nil {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}
		
		if err := userService.UpdateProfile(nowUserID, input); err != nil { // 只更新本次传了的字段
			apperror.AbortWithException(c, apperror.ServerError, err)
			return
		}

		user, err := userService.GetProfile(nowUserID) // 更新后再查一次最新资料返回给前端
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, user) // 返回更新后的用户信息
	}
	
}