// 发布帖子
// Package post 是帖子模块的 HTTP 处理器层。
// 本文件对应“发布帖子”接口。因为帖子带可选图片，所以这里用 multipart/form-data 表单接收，
// 而不是像其它接口那样用 JSON；这也是本文件中出现文件处理逻辑的原因。
package post

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"LAF/internal/middleware"
	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// Create 是“发布帖子”的处理器工厂。
// 图片只存相对路径(如 /uploads/posts/xxx.png)，因此不需要注入对外基础地址。
func Create(postService *service.PostService) gin.HandlerFunc {
	return func(c *gin.Context) {
		postType := c.PostForm("type") // 从 multipart 表单读取普通字段(注意不是 JSON)
		title := c.PostForm("title")
		content := c.PostForm("content")
		

		if postType == "" || content == "" || title == "" { // 必填字段校验：缺失直接报参数错误
			apperror.AbortWithException(c, apperror.ParamError, nil)
			return
		}

		file, header, err := c.Request.FormFile("image") // 读取可选图片；没有图片时返回 ErrMissingFile，须排除这种“正常缺省”
		if err != nil && err != http.ErrMissingFile {
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		var imageURL *string
		if file != nil {
			defer file.Close()

			if header != nil && header.Filename != "" {
				ext := filepath.Ext(header.Filename) // 取原文件扩展名，保留图片格式
				fileName := fmt.Sprintf("%d_%d%s", time.Now().UnixNano(), time.Now().Unix(), ext) // 用纳秒+秒时间戳命名，尽量保证文件名唯一，避免覆盖
				relativePath := filepath.ToSlash(filepath.Join("uploads", "posts", fileName))
				diskPath := filepath.Join(".", relativePath)
				if err := os.MkdirAll(filepath.Dir(diskPath), 0o755); err != nil { // 确保上传目录存在(不存在则创建)
					apperror.AbortWithException(c, apperror.ServerError, err)
					return
				}
				if err := c.SaveUploadedFile(header, diskPath); err != nil { // 把上传的文件写入磁盘
					apperror.AbortWithException(c, apperror.ServerError, err)
					return
				}
				url := "/" + relativePath // 只存相对路径(如 /uploads/posts/xxx.png)，由前端同源访问；避免写死主机，换 IP/域名后图片仍可访问
				imageURL = &url
			}
		}

		// 可选位置信息：location_id(手动选择)与 latitude/longitude(自动匹配)二选一，supplement 为补充说明。
		locationID := c.PostForm("location_id")
		supplement := c.PostForm("supplement")
		latitude, longitude, hasCoords, coordErr := parseOptionalCoords(c.PostForm("latitude"), c.PostForm("longitude"))
		if coordErr != nil {
			apperror.AbortWithException(c, apperror.ParamError, coordErr)
			return
		}

		nowUserID, _ := middleware.CurrentUserID(c)
		nowUserRole, _ := middleware.CurrentRole(c)
		createdPost, err := postService.Create(service.CreateInput{ // 调用业务层创建帖子(作者取登录身份，状态由业务层按角色决定)
			Type:       postType,
			Title:      title,
			Content:    content,
			ImageURL:   imageURL,
			LocationID: locationID,
			Latitude:   latitude,
			Longitude:  longitude,
			HasCoords:  hasCoords,
			Supplement: supplement,
		}, nowUserID, nowUserRole)

		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}

		response.Success(c, createdPost) // 返回创建好的帖子
	}
}

// parseOptionalCoords 解析可选的坐标字段(latitude/longitude)。
// 约束：两者要么都不传(返回 hasCoords=false，表示未上报坐标)，要么都传且都是合法数字；
// 只传其一或解析失败都视为参数错误，避免“半个坐标”被静默忽略。
func parseOptionalCoords(latitudeRaw, longitudeRaw string) (float64, float64, bool, error) {
	latStr := strings.TrimSpace(latitudeRaw)
	lonStr := strings.TrimSpace(longitudeRaw)
	if latStr == "" && lonStr == "" {
		return 0, 0, false, nil
	}
	if latStr == "" || lonStr == "" {
		return 0, 0, false, apperror.ParamError
	}
	latitude, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return 0, 0, false, apperror.ParamError
	}
	longitude, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return 0, 0, false, apperror.ParamError
	}
	return latitude, longitude, true, nil
}
