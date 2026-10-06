// 本文件对应“修改帖子完成状态”接口(管理员通用入口，可反复切换完成/未完成)。
package postadmin

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// UpdatePostFinishedRequest 是改完成状态请求体：只含目标完成标记 finished。
// 用 *bool + required 以区分“未传”与“显式传 false”，避免漏传时误把帖子设为未完成。
type UpdatePostFinishedRequest struct {
	Finished *bool `json:"finished" binding:"required"`
}

// UpdatePostFinished 是“修改帖子完成状态”的处理器工厂。
func UpdatePostFinished(postAdminService *service.PostAdminService) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID, err := strconv.ParseUint(c.Param("post_id"), 10, 64) // 解析路径中的 post_id
		if err != nil || postID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}

		var request UpdatePostFinishedRequest
		if err := c.ShouldBindJSON(&request); err != nil {
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		if err := postAdminService.UpdatePostFinished(postID, *request.Finished); err != nil { // 校验帖子存在后更新 is_finished
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, gin.H{"post_id": postID, "is_finished": *request.Finished}) // 返回帖子 id 与更新后的完成状态
	}
}