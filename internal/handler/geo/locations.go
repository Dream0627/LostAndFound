// Package geo 是“地理位置”模块的 HTTP 处理器(handler)层。
// handler 只做“解析请求参数”和“写回响应”，业务规则交给 service。
// 本文件对应“获取校园预设地点列表”接口(公开，无需登录)，供前端渲染手动选择器。
package geo

import (
	"github.com/gin-gonic/gin"

	"LAF/internal/service"
	"LAF/pkg/response"
)

// ListLocations 是“校园预设地点列表”的处理器工厂。
func ListLocations(geoService *service.GeoService) gin.HandlerFunc {
	return func(c *gin.Context) {
		response.Success(c, geoService.ListLocations()) // 返回 [{campus, locations:[...]}, ...]
	}
}
