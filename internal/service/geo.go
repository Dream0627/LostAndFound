// 本文件是“地理位置”模块的业务逻辑层：提供校园预设地点列表与“定位/匹配”能力。
// 它不依赖数据库，只封装校验规则；坐标来源(前端上报)由调用方负责。
package service

import (
	"strings"

	"LAF/pkg/apperror"
	"LAF/pkg/geo"
)

// GeoService 提供地理位置相关业务能力。无外部依赖，故构造函数无参数。
type GeoService struct{}

// NewGeoService 构造 GeoService。
func NewGeoService() *GeoService {
	return &GeoService{}
}

// LocationGroup 是按校区分组后的地点，便于前端直接渲染成分组选择器。
type LocationGroup struct {
	Campus    string         `json:"campus"`
	Locations []geo.Location `json:"locations"`
}

// ListLocations 返回全部校区的地点，按校区聚合成组。
// 在服务层分组而不是让前端分组：前端只需“校区 -> 地点”两级结构，减少页面逻辑。
func (s *GeoService) ListLocations() []LocationGroup {
	order := []string{geo.CampusPingfeng, geo.CampusZhaohui, geo.CampusMoganshan}
	result := make([]LocationGroup, 0, len(order))
	for _, campus := range order {
		locations := geo.LocationsByCampus(campus)
		if len(locations) == 0 {
			continue
		}
		result = append(result, LocationGroup{Campus: campus, Locations: locations})
	}
	return result
}

// LocateInput 是“定位/匹配”的入参 DTO。
// 两种用法二选一：
//  1) 手动选择：传 LocationID；
//  2) 自动定位：传 Latitude/Longitude(并把 HasCoords 置真)，由后端匹配最近的预设地点。
// Supplement 是可选补充说明(手动选择后可再补充，如“图书馆东门台阶旁”)。
type LocateInput struct {
	LocationID string
	Latitude   float64
	Longitude  float64
	Supplement string
	HasCoords  bool // 调用方是否真的上报了坐标(用于区分“未上报”与“(0,0)”)
}

// LocateResult 是“定位/匹配”的出参 DTO。
type LocateResult struct {
	Location       geo.Location `json:"location"`        // 匹配/选择到的预设地点
	DistanceMeters float64      `json:"distance_meters"` // 与预设地点的直线距离(自动定位时有意义)
	MatchType      string       `json:"match_type"`      // manual=手动选择, auto=自动匹配
	Supplement     string       `json:"supplement"`      // 用户补充说明
}

// maxSupplementLength 是补充说明的最大长度(按字符数)，防止超长文本。
const maxSupplementLength = 200

// Locate 执行定位/匹配并校验入参，返回最终地点与补充说明。
// 校验顺序：先校验补充说明长度 -> 优先按手动选择的地点 ID 解析 -> 再按上报坐标自动匹配 -> 都没有则报参数错误。
func (s *GeoService) Locate(input LocateInput) (*LocateResult, error) {
	supplement := strings.TrimSpace(input.Supplement)
	if len([]rune(supplement)) > maxSupplementLength {
		return nil, apperror.ParamError
	}

	// 优先“手动选择”。
	if id := strings.TrimSpace(input.LocationID); id != "" {
		loc, ok := geo.FindByID(id)
		if !ok {
			return nil, apperror.GeoLocationNotFoundError
		}
		return &LocateResult{Location: loc, MatchType: "manual", Supplement: supplement}, nil
	}

	// 其次“自动定位匹配最近地点”。
	if input.HasCoords {
		coords := geo.Coordinates{Latitude: input.Latitude, Longitude: input.Longitude}
		if !geo.IsValidCoordinates(coords) {
			return nil, apperror.InvalidCoordinateError
		}
		loc, dist, ok := geo.Nearest(coords)
		if !ok {
			return nil, apperror.GeoLocationNotFoundError
		}
		return &LocateResult{Location: loc, DistanceMeters: dist, MatchType: "auto", Supplement: supplement}, nil
	}

	// 既没选地点也没上报坐标：参数不完整。
	return nil, apperror.ParamError
}
