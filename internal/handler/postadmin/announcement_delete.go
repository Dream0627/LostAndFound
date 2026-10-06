package postadmin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// DeleteAnnouncement 是“删除公告”的处理器工厂。
// 本文件对应“删除公告”接口。
// 流程：解析路径参数 -> 读取当前用户身份 -> 校验管理员权限 -> 执行软删除。
func DeleteAnnouncement(announcementService *service.AnnouncementService) gin.HandlerFunc {
	return func(c *gin.Context) {
		announcementID, err := strconv.ParseUint(c.Param("announcement_id"), 10, 64) // 把路径里的 announcement_id 从字符串解析成无符号整数(10 进制, 64 位)
		if err != nil || announcementID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}
		nowUserRole, _ := middleware.CurrentRole(c)                   // 取当前角色，用于后面的权限判断
		if nowUserRole != "postadmin" && nowUserRole != "mainadmin" { // 只有管理员可以删除公告
			apperror.AbortWithError(c, apperror.AdminForbiddenError)
			return
		}
		if err := announcementService.DeleteAnnouncement(announcementID); err != nil {
			apperror.AbortWithError(c, err)
			return
		}
		response.Success(c, nil) // 删除成功返回空数据
	}
}
