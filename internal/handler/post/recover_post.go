package post

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

func RecoverPost(postService *service.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		postID, err := strconv.ParseUint(c.Param("post_id"), 10, 64)
		if err != nil || postID == 0 {
			response.Error(c, apperror.ParamError.Code, apperror.ParamError.Msg)
			return
		}

		nowUserID, _ := middleware.CurrentUserID(c)
		nowUserRole, _ := middleware.CurrentRole(c)
		post, err := postService.GetPostByIDUnscoped(postID)

		if err != nil {
			if errors.Is(err, apperror.NotFoundError) || err == apperror.NotFoundError {
				apperror.AbortWithException(c, apperror.NotFoundError, nil)
			} else {
				apperror.AbortWithError(c, err)
			}
			return
		}

		if err := postService.CheckpostPermission(nowUserID, nowUserRole, post); err != nil {
			apperror.AbortWithException(c, apperror.UserForbiddenError, nil)
		}

		if err := postService.RecoverPost(postID); err != nil {
			apperror.AbortWithError(c, err)
		}
		response.Success(c, gin.H{"post_id": postID})
	}

}
