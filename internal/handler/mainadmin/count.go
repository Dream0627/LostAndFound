package mainadmin

import (
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"

	"github.com/gin-gonic/gin"
)

// GetCountHandler 是“后台数据计数”的处理器工厂。
// 本文件对应 GET /api/v1/admin/count 接口，仅超级管理员可访问。
func GetCountHandler(mainAdminService *service.MainAdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		count, err := mainAdminService.GetCount()
		if err != nil {
			apperror.AbortWithException(c, apperror.DatabaseError, err)
			return
		}
		response.Success(c, count)
	}
}
