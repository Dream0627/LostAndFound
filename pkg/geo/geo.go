// Package geo 提供“用户地理位置”相关的通用工具。
// 设计目标(对应选型方案 E+D)：
//  1) 采用“客户端上报坐标 + 校园预设地点”的方案，核心逻辑只依赖标准库，不引入任何第三方依赖；
//  2) 只定义数据结构与纯函数(坐标距离、最近地点匹配、客户端 IP 提取)，
//     具体如何取得坐标由前端/调用方决定，便于替换与测试；
//  3) 校内地点有限且固定，用一张静态“校园地点表”(见 campus.go)即可把坐标翻译成可读地名，
//     无需在线逆地理 API，也无需离线 mmdb 库。
package geo

import (
	"math"
	"strings"

	"github.com/gin-gonic/gin"
)

// Coordinates 表示一个经纬度坐标点，单位：度。
// 采用 WGS-84 坐标系(浏览器 Geolocation API 默认返回 WGS-84)，前端可直接上报。
type Coordinates struct {
	Latitude  float64 `json:"latitude"`  // 纬度
	Longitude float64 `json:"longitude"` // 经度
}

// Location 表示一个校园预设地点(楼栋/店铺等)。
// ID 是稳定的字符串标识，前端选择与后端匹配都以它为准；
// Campus 用于按校区分组展示；Category 用于区分“教学楼/食堂/宿舍”等类别。
type Location struct {
	ID        string  `json:"id"`        // 地点唯一标识(如 "pf-library")
	Campus    string  `json:"campus"`    // 所属校区
	Name      string  `json:"name"`      // 地点名称(如 "图书馆")
	Category  string  `json:"category"`  // 类别(教学楼/图书馆/食堂/宿舍/运动场馆/生活服务/其他)
	Address   string  `json:"address"`   // 补充描述(如 "屏峰·中心区")
	Latitude  float64 `json:"latitude"`  // 纬度(近似值)
	Longitude float64 `json:"longitude"` // 经度(近似值)
}

// Coordinates 返回该地点的坐标点，便于与 Distance 等函数配合使用。
func (l Location) Coordinates() Coordinates {
	return Coordinates{Latitude: l.Latitude, Longitude: l.Longitude}
}

// earthRadiusMeters 是地球平均半径，单位：米，用于 Haversine 公式。
const earthRadiusMeters = 6371000.0

// Distance 用 Haversine(半正矢)公式计算两个经纬度点之间的球面距离，单位：米。
// 之所以不用简单勾股：经度在球面上随纬度收窄，直接线性相减误差很大；
// Haversine 在校园尺度上足够精确，且只需标准库 math。
func Distance(a, b Coordinates) float64 {
	lat1 := a.Latitude * math.Pi / 180
	lat2 := b.Latitude * math.Pi / 180
	dLat := (b.Latitude - a.Latitude) * math.Pi / 180
	dLon := (b.Longitude - a.Longitude) * math.Pi / 180

	sinDLat := math.Sin(dLat / 2)
	sinDLon := math.Sin(dLon / 2)
	h := sinDLat*sinDLat + math.Cos(lat1)*math.Cos(lat2)*sinDLon*sinDLon
	return earthRadiusMeters * 2 * math.Atan2(math.Sqrt(h), math.Sqrt(1-h))
}

// IsValidCoordinates 判断坐标是否落在合法经纬度范围内。
// 纬度 [-90,90]，经度 [-180,180]；(0,0) 视为“未上报/无效”。
func IsValidCoordinates(c Coordinates) bool {
	if c.Latitude == 0 && c.Longitude == 0 {
		return false
	}
	if c.Latitude < -90 || c.Latitude > 90 {
		return false
	}
	if c.Longitude < -180 || c.Longitude > 180 {
		return false
	}
	return true
}

// FindByID 在全部校园预设地点中按 ID 查找；found 为 false 表示不存在。
func FindByID(id string) (Location, bool) {
	for _, loc := range CampusLocations() {
		if loc.ID == id {
			return loc, true
		}
	}
	return Location{}, false
}

// Nearest 在全部校园预设地点中找出离给定坐标最近的一个，并返回距离(米)。
// 地点表为空时 found 返回 false。
func Nearest(c Coordinates) (Location, float64, bool) {
	locations := CampusLocations()
	if len(locations) == 0 {
		return Location{}, 0, false
	}
	best := locations[0]
	bestDist := Distance(c, best.Coordinates())
	for _, loc := range locations[1:] {
		d := Distance(c, loc.Coordinates())
		if d < bestDist {
			best, bestDist = loc, d
		}
	}
	return best, bestDist, true
}

// ClientIP 从请求中提取客户端 IP，作为“定位兜底”使用(如用户拒绝授权时按 IP 粗略判断所在校区)。
// 取值优先级：X-Forwarded-For 第一个地址 -> X-Real-IP -> gin 内置解析(含 RemoteAddr)。
// 注意：请求头可被伪造，只能用于粗粒度兜底，不能当作可信身份。
func ClientIP(c *gin.Context) string {
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		if ip := strings.TrimSpace(parts[0]); ip != "" {
			return ip
		}
	}
	if xri := strings.TrimSpace(c.GetHeader("X-Real-IP")); xri != "" {
		return xri
	}
	return c.ClientIP()
}
