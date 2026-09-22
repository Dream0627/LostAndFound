// 本文件对应“定位/匹配最近地点”接口(公开，无需登录)。
// 前端把 GPS 坐标或手动选择的地点 ID 提交上来，后端回填可读地点名与距离。
package geo

import (
	"github.com/gin-gonic/gin"

	"LAF/internal/service"
	"LAF/pkg/apperror"
	"LAF/pkg/response"
)

// LocateRequest 是“定位/匹配”的请求体。
// location_id 与 latitude/longitude 二选一：前者表示手动选择，后者表示自动上报坐标。
// 用 *float64 指针是为了区分“字段未传”与“传了 0”。
type LocateRequest struct {
	LocationID string   `json:"location_id"`
	Latitude   *float64 `json:"latitude"`
	Longitude  *float64 `json:"longitude"`
	Supplement string   `json:"supplement"`
}

// Locate 是“定位/匹配”的处理器工厂。
func Locate(geoService *service.GeoService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var request LocateRequest
		if err := c.ShouldBindJSON(&request); err != nil { // 解析请求体；失败按参数错误处理
			apperror.AbortWithException(c, apperror.ParamError, err)
			return
		}

		input := service.LocateInput{LocationID: request.LocationID, Supplement: request.Supplement}
		if request.Latitude != nil && request.Longitude != nil { // 仅当经纬度都传入时才算“上报了坐标”
			input.Latitude = *request.Latitude
			input.Longitude = *request.Longitude
			input.HasCoords = true
		}

		result, err := geoService.Locate(input)
		if err != nil {
			apperror.AbortWithError(c, err)
			return
		}
		response.Success(c, result)
	}
}
