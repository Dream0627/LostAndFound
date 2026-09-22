// 本文件对应超级管理员“恢复用户”接口。
package mainadmin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// RecoverUser 是“恢复用户”的处理器工厂。
func RecoverUser(mainAdminService *service.MainAdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := strconv.ParseUint(c.Param("user_id"), 10, 64) // 解析路径中的 user_id
		if err != nil || userID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}

		if err := mainAdminService.RecoverUser(userID); err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, gin.H{"user_id": userID})
	}
}
