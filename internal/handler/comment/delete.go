// 删除评论
// 本文件对应“删除评论”接口。
// 流程：解析路径参数 -> 读取当前用户身份 -> 查评论 -> 校验权限 -> 执行删除。
package comment

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// Delete 是“删除评论”的处理器工厂。
func Delete(commentService *service.CommentService) gin.HandlerFunc {
	return func(c *gin.Context) {
		commentID, err := strconv.ParseUint(c.Param("comment_id"), 10, 64) // 把路径里的 comment_id 从字符串解析成无符号整数(10 进制, 64 位)
		if err != nil || commentID == 0 {
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		nowUserID, _ := middleware.CurrentUserID(c)
		nowUserRole, _ := middleware.CurrentRole(c) // 取当前角色，用于后面的权限判断

		comment, err := commentService.GetCommentByID(commentID)
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		if err := commentService.CheckCommentPermission(nowUserID, nowUserRole, comment); err != nil { // 权限校验：只有本人或管理员可删
			apperror.AbortWithError(c, err)
			return
		}

		if err := commentService.DeleteComment(commentID); err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, nil) // 删除成功返回空数据
	}
}
