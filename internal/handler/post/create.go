// 发布帖子
package post

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// 构建图片URL(绝对路径,去除前后可能存在的斜杠"/")
func buildPublicImageURL(publicBaseURL, relativePath string) string {
	base := strings.TrimRight(publicBaseURL, "/")
	path := strings.TrimLeft(relativePath, "/")
	if base == "" {
		return path
	}
	if path == "" {
		return base
	}
	return fmt.Sprintf("%s/%s", base, path)
}

//处理发布帖子数据
func Create(postService *service.PostService, publicBaseURL string) gin.HandlerFunc {
	return func(c *gin.Context) {
		postType := c.PostForm("type")
		content := c.PostForm("content")

		if postType == "" || content == "" {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}

		file, header, err := c.Request.FormFile("image")
		if err != nil && err != http.ErrMissingFile {
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}

		var imageURL *string
		if file != nil {
			defer file.Close()

			if header != nil && header.Filename != "" {
				ext := filepath.Ext(header.Filename)
				fileName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), time.Now().Unix(), ext)
				relativePath := filepath.ToSlash(filepath.Join("uploads", "posts", fileName))
				diskPath := filepath.Join(".", relativePath)
				if err := os.MkdirAll(filepath.Dir(diskPath), 0o755); err != nil {
					apperror.AbortWithException(c, apperror.ServerError, err)
					return
				}
				if err := c.SaveUploadedFile(header, diskPath); err != nil {
					apperror.AbortWithException(c, apperror.ServerError, err)
					return
				}
				url := buildPublicImageURL(publicBaseURL, relativePath)
				imageURL = &url
			}
		}

		userID := c.GetUint(middleware.UserIDKey)
		createdPost, err := postService.Create(service.CreateInput{
			Type:     postType,
			Content:  content,
			ImageURL: imageURL,
		}, userID)

		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, createdPost)
	}
}
