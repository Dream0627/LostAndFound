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

// 国内地图(高德/腾讯)使用 GCJ-02(火星坐标系)，而浏览器 Geolocation 按规范返回 WGS-84，
// 杭州一带两者相差约 400~600 米——足以让“匹配最近地点”跨片区选错楼。
// 下面用国测局公开的偏移算法把 WGS-84 转成 GCJ-02，只依赖标准库 math，不引入第三方依赖。
const (
	gcjAxis  = 6378245.0              // 克拉索夫斯基椭球长半轴(米)
	gcjEccSq = 0.00669342162296594323 // 第一偏心率的平方
)

// outOfChina 判断坐标是否在中国大陆范围外；境外无 GCJ-02 偏移，应保持原坐标。
func outOfChina(lat, lon float64) bool {
	return lon < 72.004 || lon > 137.8347 || lat < 0.8293 || lat > 55.8271
}

// gcjOffsetLat / gcjOffsetLon 是偏移量的中间多项式(国测局公开算法)，x 为经度差、y 为纬度差。
func gcjOffsetLat(x, y float64) float64 {
	ret := -100 + 2*x + 3*y + 0.2*y*y + 0.1*x*y + 0.2*math.Sqrt(math.Abs(x))
	ret += (20*math.Sin(6*x*math.Pi) + 20*math.Sin(2*x*math.Pi)) * 2 / 3
	ret += (20*math.Sin(y*math.Pi) + 40*math.Sin(y/3*math.Pi)) * 2 / 3
	ret += (160*math.Sin(y/12*math.Pi) + 320*math.Sin(y*math.Pi/30)) * 2 / 3
	return ret
}

func gcjOffsetLon(x, y float64) float64 {
	ret := 300 + x + 2*y + 0.1*x*x + 0.1*x*y + 0.1*math.Sqrt(math.Abs(x))
	ret += (20*math.Sin(6*x*math.Pi) + 20*math.Sin(2*x*math.Pi)) * 2 / 3
	ret += (20*math.Sin(x*math.Pi) + 40*math.Sin(x/3*math.Pi)) * 2 / 3
	ret += (150*math.Sin(x/12*math.Pi) + 300*math.Sin(x/30*math.Pi)) * 2 / 3
	return ret
}

// WGSToGCJ02 把 WGS-84 坐标转换成 GCJ-02。
// 用途：浏览器定位返回的是 WGS-84，而校园预设地点表(见 campus.go)用的是 GCJ-02，
// 匹配前必须先统一坐标系，否则会出现数百米的系统性偏差。
func WGSToGCJ02(c Coordinates) Coordinates {
	if (c.Latitude == 0 && c.Longitude == 0) || outOfChina(c.Latitude, c.Longitude) {
		return c
	}
	dLat := gcjOffsetLat(c.Longitude-105, c.Latitude-35)
	dLon := gcjOffsetLon(c.Longitude-105, c.Latitude-35)
	radLat := c.Latitude / 180 * math.Pi
	magic := 1 - gcjEccSq*math.Sin(radLat)*math.Sin(radLat)
	sqrtMagic := math.Sqrt(magic)
	dLat = dLat * 180 / ((gcjAxis * (1 - gcjEccSq)) / (magic * sqrtMagic) * math.Pi)
	dLon = dLon * 180 / (gcjAxis / sqrtMagic * math.Cos(radLat) * math.Pi)
	return Coordinates{Latitude: c.Latitude + dLat, Longitude: c.Longitude + dLon}
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
