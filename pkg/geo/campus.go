// 本文件定义浙江工业大学三校区的“校园预设地点表”。
// 说明：
//   - 宿舍按“每一栋楼”单独列出(如“家和东苑3号楼”“尚德园5号楼”“德馨苑·德三楼”)，便于精确到楼；
//   - 坐标为近似值，仅用于距离比较与校区归属判断，校园尺度足够；上线前请以学校官方校园地图做一次校准；
//   - 地点名称依据公开资料整理，供“手动选择”兜底使用；
//   - 新增/调整地点只需增删下方 campusLocations 条目，其余代码无需改动。
package geo

// 校区名称常量，避免在多处手写中文字符串导致不一致。
const (
	CampusPingfeng  = "屏峰校区"
	CampusZhaohui   = "朝晖校区"
	CampusMoganshan = "莫干山校区"
)

// 地点类别常量。
const (
	CategoryTeaching = "教学楼"
	CategoryLibrary  = "图书馆"
	CategoryDining   = "食堂"
	CategoryDorm     = "宿舍"
	CategorySport    = "运动场馆"
	CategoryService  = "生活服务"
	CategoryOther    = "其他"
)

// CampusLocations 返回全部校园预设地点(每次返回副本，避免调用方误改全局数据)。
// 使用函数而非包级导出变量，是为了保护内部数据，也便于将来改为从数据库/配置文件加载。
func CampusLocations() []Location {
	out := make([]Location, len(campusLocations))
	copy(out, campusLocations)
	return out
}

// LocationsByCampus 返回指定校区的地点列表；campus 为空时返回全部。
func LocationsByCampus(campus string) []Location {
	if campus == "" {
		return CampusLocations()
	}
	var out []Location
	for _, loc := range campusLocations {
		if loc.Campus == campus {
			out = append(out, loc)
		}
	}
	return out
}

// campusLocations 是内置的预设地点数据(包级私有)。
// 宿舍已细化到每一栋楼(家和东苑1-18、家和西苑1-15、尚德园1-9、梦溪村1-7、综合楼、德馨苑德一~德十)。
var campusLocations = []Location{
	{ID: "pf-library", Campus: CampusPingfeng, Name: "图书馆", Category: CategoryLibrary, Address: "屏峰·中心区", Latitude: 30.23390, Longitude: 120.09310},
	{ID: "pf-jianxing", Campus: CampusPingfeng, Name: "健行楼", Category: CategoryTeaching, Address: "屏峰·教学区", Latitude: 30.23420, Longitude: 120.09280},
	{ID: "pf-guangzhi", Campus: CampusPingfeng, Name: "广知楼", Category: CategoryTeaching, Address: "屏峰·教学区", Latitude: 30.23370, Longitude: 120.09250},
	{ID: "pf-yulin", Campus: CampusPingfeng, Name: "语林楼", Category: CategoryTeaching, Address: "屏峰·教学区", Latitude: 30.23340, Longitude: 120.09290},
	{ID: "pf-lixue", Campus: CampusPingfeng, Name: "理学楼", Category: CategoryTeaching, Address: "屏峰·教学区", Latitude: 30.23310, Longitude: 120.09330},
	{ID: "pf-boyi", Campus: CampusPingfeng, Name: "博易楼", Category: CategoryTeaching, Address: "屏峰·教学区", Latitude: 30.23400, Longitude: 120.09360},
	{ID: "pf-changyuan", Campus: CampusPingfeng, Name: "畅远楼", Category: CategoryTeaching, Address: "屏峰·教学区", Latitude: 30.23440, Longitude: 120.09320},
	{ID: "pf-faxue", Campus: CampusPingfeng, Name: "法学楼", Category: CategoryTeaching, Address: "屏峰·教学区", Latitude: 30.23320, Longitude: 120.09380},
	{ID: "pf-renhe", Campus: CampusPingfeng, Name: "仁和楼", Category: CategoryTeaching, Address: "屏峰·教学区", Latitude: 30.23380, Longitude: 120.09400},
	{ID: "pf-yuwen", Campus: CampusPingfeng, Name: "郁文楼", Category: CategoryTeaching, Address: "屏峰·教学区", Latitude: 30.23430, Longitude: 120.09370},
	{ID: "pf-yangxian", Campus: CampusPingfeng, Name: "养贤府食堂", Category: CategoryDining, Address: "屏峰·生活区", Latitude: 30.23280, Longitude: 120.09240},
	{ID: "pf-jiahe-canteen", Campus: CampusPingfeng, Name: "家和食堂", Category: CategoryDining, Address: "屏峰·生活区", Latitude: 30.23260, Longitude: 120.09410},
	{ID: "pf-gym", Campus: CampusPingfeng, Name: "体育馆", Category: CategorySport, Address: "屏峰·文体区", Latitude: 30.23500, Longitude: 120.09290},
	{ID: "pf-track", Campus: CampusPingfeng, Name: "田径场", Category: CategorySport, Address: "屏峰·文体区", Latitude: 30.23530, Longitude: 120.09330},
	{ID: "pf-service", Campus: CampusPingfeng, Name: "师生服务中心", Category: CategoryService, Address: "屏峰·生活区", Latitude: 30.23300, Longitude: 120.09220},
	{ID: "pf-sunflower", Campus: CampusPingfeng, Name: "向日葵花海", Category: CategoryOther, Address: "屏峰·景观区", Latitude: 30.23480, Longitude: 120.09430},
	{ID: "pf-jiahe-east-1", Campus: CampusPingfeng, Name: "家和东苑1号楼", Category: CategoryDorm, Address: "屏峰·家和东苑", Latitude: 30.23240, Longitude: 120.09360},
	{ID: "pf-jiahe-east-2", Campus: CampusPingfeng, Name: "家和东苑2号楼", Category: CategoryDorm, Address: "屏峰·家和东苑", Latitude: 30.23247, Longitude: 120.09360},
	{ID: "pf-jiahe-east-3", Campus: CampusPingfeng, Name: "家和东苑3号楼", Category: CategoryDorm, Address: "屏峰·家和东苑", Latitude: 30.23254, Longitude: 120.09360},
	{ID: "pf-jiahe-east-4", Campus: CampusPingfeng, Name: "家和东苑4号楼", Category: CategoryDorm, Address: "屏峰·家和东苑", Latitude: 30.23261, Longitude: 120.09360},
	{ID: "pf-jiahe-east-5", Campus: CampusPingfeng, Name: "家和东苑5号楼", Category: CategoryDorm, Address: "屏峰·家和东苑", Latitude: 30.23268, Longitude: 120.09360},
	{ID: "pf-jiahe-east-6", Campus: CampusPingfeng, Name: "家和东苑6号楼", Category: CategoryDorm, Address: "屏峰·家和东苑", Latitude: 30.23240, Longitude: 120.09367},
	{ID: "pf-jiahe-east-7", Campus: CampusPingfeng, Name: "家和东苑7号楼", Category: CategoryDorm, Address: "屏峰·家和东苑", Latitude: 30.23247, Longitude: 120.09367},
	{ID: "pf-jiahe-east-8", Campus: CampusPingfeng, Name: "家和东苑8号楼", Category: CategoryDorm, Address: "屏峰·家和东苑", Latitude: 30.23254, Longitude: 120.09367},
	{ID: "pf-jiahe-east-9", Campus: CampusPingfeng, Name: "家和东苑9号楼", Category: CategoryDorm, Address: "屏峰·家和东苑", Latitude: 30.23261, Longitude: 120.09367},
	{ID: "pf-jiahe-east-10", Campus: CampusPingfeng, Name: "家和东苑10号楼", Category: CategoryDorm, Address: "屏峰·家和东苑", Latitude: 30.23268, Longitude: 120.09367},
	{ID: "pf-jiahe-east-11", Campus: CampusPingfeng, Name: "家和东苑11号楼", Category: CategoryDorm, Address: "屏峰·家和东苑", Latitude: 30.23240, Longitude: 120.09374},
	{ID: "pf-jiahe-east-12", Campus: CampusPingfeng, Name: "家和东苑12号楼", Category: CategoryDorm, Address: "屏峰·家和东苑", Latitude: 30.23247, Longitude: 120.09374},
	{ID: "pf-jiahe-east-13", Campus: CampusPingfeng, Name: "家和东苑13号楼", Category: CategoryDorm, Address: "屏峰·家和东苑", Latitude: 30.23254, Longitude: 120.09374},
	{ID: "pf-jiahe-east-14", Campus: CampusPingfeng, Name: "家和东苑14号楼", Category: CategoryDorm, Address: "屏峰·家和东苑", Latitude: 30.23261, Longitude: 120.09374},
	{ID: "pf-jiahe-east-15", Campus: CampusPingfeng, Name: "家和东苑15号楼", Category: CategoryDorm, Address: "屏峰·家和东苑", Latitude: 30.23268, Longitude: 120.09374},
	{ID: "pf-jiahe-east-16", Campus: CampusPingfeng, Name: "家和东苑16号楼", Category: CategoryDorm, Address: "屏峰·家和东苑", Latitude: 30.23240, Longitude: 120.09381},
	{ID: "pf-jiahe-east-17", Campus: CampusPingfeng, Name: "家和东苑17号楼", Category: CategoryDorm, Address: "屏峰·家和东苑", Latitude: 30.23247, Longitude: 120.09381},
	{ID: "pf-jiahe-east-18", Campus: CampusPingfeng, Name: "家和东苑18号楼", Category: CategoryDorm, Address: "屏峰·家和东苑", Latitude: 30.23254, Longitude: 120.09381},
	{ID: "pf-jiahe-west-1", Campus: CampusPingfeng, Name: "家和西苑1号楼", Category: CategoryDorm, Address: "屏峰·家和西苑", Latitude: 30.23220, Longitude: 120.09270},
	{ID: "pf-jiahe-west-2", Campus: CampusPingfeng, Name: "家和西苑2号楼", Category: CategoryDorm, Address: "屏峰·家和西苑", Latitude: 30.23227, Longitude: 120.09270},
	{ID: "pf-jiahe-west-3", Campus: CampusPingfeng, Name: "家和西苑3号楼", Category: CategoryDorm, Address: "屏峰·家和西苑", Latitude: 30.23234, Longitude: 120.09270},
	{ID: "pf-jiahe-west-4", Campus: CampusPingfeng, Name: "家和西苑4号楼", Category: CategoryDorm, Address: "屏峰·家和西苑", Latitude: 30.23241, Longitude: 120.09270},
	{ID: "pf-jiahe-west-5", Campus: CampusPingfeng, Name: "家和西苑5号楼", Category: CategoryDorm, Address: "屏峰·家和西苑", Latitude: 30.23248, Longitude: 120.09270},
	{ID: "pf-jiahe-west-6", Campus: CampusPingfeng, Name: "家和西苑6号楼", Category: CategoryDorm, Address: "屏峰·家和西苑", Latitude: 30.23220, Longitude: 120.09277},
	{ID: "pf-jiahe-west-7", Campus: CampusPingfeng, Name: "家和西苑7号楼", Category: CategoryDorm, Address: "屏峰·家和西苑", Latitude: 30.23227, Longitude: 120.09277},
	{ID: "pf-jiahe-west-8", Campus: CampusPingfeng, Name: "家和西苑8号楼", Category: CategoryDorm, Address: "屏峰·家和西苑", Latitude: 30.23234, Longitude: 120.09277},
	{ID: "pf-jiahe-west-9", Campus: CampusPingfeng, Name: "家和西苑9号楼", Category: CategoryDorm, Address: "屏峰·家和西苑", Latitude: 30.23241, Longitude: 120.09277},
	{ID: "pf-jiahe-west-10", Campus: CampusPingfeng, Name: "家和西苑10号楼", Category: CategoryDorm, Address: "屏峰·家和西苑", Latitude: 30.23248, Longitude: 120.09277},
	{ID: "pf-jiahe-west-11", Campus: CampusPingfeng, Name: "家和西苑11号楼", Category: CategoryDorm, Address: "屏峰·家和西苑", Latitude: 30.23220, Longitude: 120.09284},
	{ID: "pf-jiahe-west-12", Campus: CampusPingfeng, Name: "家和西苑12号楼", Category: CategoryDorm, Address: "屏峰·家和西苑", Latitude: 30.23227, Longitude: 120.09284},
	{ID: "pf-jiahe-west-13", Campus: CampusPingfeng, Name: "家和西苑13号楼", Category: CategoryDorm, Address: "屏峰·家和西苑", Latitude: 30.23234, Longitude: 120.09284},
	{ID: "pf-jiahe-west-14", Campus: CampusPingfeng, Name: "家和西苑14号楼", Category: CategoryDorm, Address: "屏峰·家和西苑", Latitude: 30.23241, Longitude: 120.09284},
	{ID: "pf-jiahe-west-15", Campus: CampusPingfeng, Name: "家和西苑15号楼", Category: CategoryDorm, Address: "屏峰·家和西苑", Latitude: 30.23248, Longitude: 120.09284},
	{ID: "zh-library", Campus: CampusZhaohui, Name: "图书馆", Category: CategoryLibrary, Address: "朝晖·中心区", Latitude: 30.29770, Longitude: 120.16550},
	{ID: "zh-ziliang", Campus: CampusZhaohui, Name: "子良楼", Category: CategoryTeaching, Address: "朝晖·教学区", Latitude: 30.29800, Longitude: 120.16580},
	{ID: "zh-xinjiao", Campus: CampusZhaohui, Name: "新教楼", Category: CategoryTeaching, Address: "朝晖·教学区", Latitude: 30.29740, Longitude: 120.16520},
	{ID: "zh-wenhui", Campus: CampusZhaohui, Name: "文荟楼", Category: CategoryTeaching, Address: "朝晖·教学区", Latitude: 30.29790, Longitude: 120.16510},
	{ID: "zh-yuxiu", Campus: CampusZhaohui, Name: "毓秀堂食堂", Category: CategoryDining, Address: "朝晖·生活区", Latitude: 30.29730, Longitude: 120.16600},
	{ID: "zh-track", Campus: CampusZhaohui, Name: "田径场", Category: CategorySport, Address: "朝晖·文体区", Latitude: 30.29850, Longitude: 120.16620},
	{ID: "zh-gym", Campus: CampusZhaohui, Name: "体育馆", Category: CategorySport, Address: "朝晖·文体区", Latitude: 30.29880, Longitude: 120.16590},
	{ID: "zh-quanjia", Campus: CampusZhaohui, Name: "全家超市(朝晖店)", Category: CategoryService, Address: "朝晖·生活区", Latitude: 30.29720, Longitude: 120.16560},
	{ID: "zh-yiming", Campus: CampusZhaohui, Name: "一鸣奶吧(朝晖店)", Category: CategoryService, Address: "朝晖·生活区", Latitude: 30.29710, Longitude: 120.16540},
	{ID: "zh-luckin", Campus: CampusZhaohui, Name: "瑞幸咖啡(朝晖店)", Category: CategoryService, Address: "朝晖·生活区", Latitude: 30.29750, Longitude: 120.16500},
	{ID: "zh-print", Campus: CampusZhaohui, Name: "打印店(朝晖)", Category: CategoryService, Address: "朝晖·生活区", Latitude: 30.29735, Longitude: 120.16575},
	{ID: "zh-glasses", Campus: CampusZhaohui, Name: "眼镜店(朝晖)", Category: CategoryService, Address: "朝晖·生活区", Latitude: 30.29725, Longitude: 120.16545},
	{ID: "zh-fruit", Campus: CampusZhaohui, Name: "水果店(朝晖)", Category: CategoryService, Address: "朝晖·生活区", Latitude: 30.29715, Longitude: 120.16515},
	{ID: "zh-barber", Campus: CampusZhaohui, Name: "理发店(朝晖)", Category: CategoryService, Address: "朝晖·生活区", Latitude: 30.29745, Longitude: 120.16565},
	{ID: "zh-shangde-1", Campus: CampusZhaohui, Name: "尚德园1号楼", Category: CategoryDorm, Address: "朝晖·尚德园", Latitude: 30.29830, Longitude: 120.16570},
	{ID: "zh-shangde-2", Campus: CampusZhaohui, Name: "尚德园2号楼", Category: CategoryDorm, Address: "朝晖·尚德园", Latitude: 30.29837, Longitude: 120.16570},
	{ID: "zh-shangde-3", Campus: CampusZhaohui, Name: "尚德园3号楼", Category: CategoryDorm, Address: "朝晖·尚德园", Latitude: 30.29844, Longitude: 120.16570},
	{ID: "zh-shangde-4", Campus: CampusZhaohui, Name: "尚德园4号楼", Category: CategoryDorm, Address: "朝晖·尚德园", Latitude: 30.29851, Longitude: 120.16570},
	{ID: "zh-shangde-5", Campus: CampusZhaohui, Name: "尚德园5号楼", Category: CategoryDorm, Address: "朝晖·尚德园", Latitude: 30.29858, Longitude: 120.16570},
	{ID: "zh-shangde-6", Campus: CampusZhaohui, Name: "尚德园6号楼", Category: CategoryDorm, Address: "朝晖·尚德园", Latitude: 30.29830, Longitude: 120.16577},
	{ID: "zh-shangde-7", Campus: CampusZhaohui, Name: "尚德园7号楼", Category: CategoryDorm, Address: "朝晖·尚德园", Latitude: 30.29837, Longitude: 120.16577},
	{ID: "zh-shangde-8", Campus: CampusZhaohui, Name: "尚德园8号楼", Category: CategoryDorm, Address: "朝晖·尚德园", Latitude: 30.29844, Longitude: 120.16577},
	{ID: "zh-shangde-9", Campus: CampusZhaohui, Name: "尚德园9号楼", Category: CategoryDorm, Address: "朝晖·尚德园", Latitude: 30.29851, Longitude: 120.16577},
	{ID: "zh-mengxi-1", Campus: CampusZhaohui, Name: "梦溪村1号楼", Category: CategoryDorm, Address: "朝晖·梦溪村", Latitude: 30.29700, Longitude: 120.16490},
	{ID: "zh-mengxi-2", Campus: CampusZhaohui, Name: "梦溪村2号楼", Category: CategoryDorm, Address: "朝晖·梦溪村", Latitude: 30.29707, Longitude: 120.16490},
	{ID: "zh-mengxi-3", Campus: CampusZhaohui, Name: "梦溪村3号楼", Category: CategoryDorm, Address: "朝晖·梦溪村", Latitude: 30.29714, Longitude: 120.16490},
	{ID: "zh-mengxi-4", Campus: CampusZhaohui, Name: "梦溪村4号楼", Category: CategoryDorm, Address: "朝晖·梦溪村", Latitude: 30.29721, Longitude: 120.16490},
	{ID: "zh-mengxi-5", Campus: CampusZhaohui, Name: "梦溪村5号楼", Category: CategoryDorm, Address: "朝晖·梦溪村", Latitude: 30.29728, Longitude: 120.16490},
	{ID: "zh-mengxi-6", Campus: CampusZhaohui, Name: "梦溪村6号楼", Category: CategoryDorm, Address: "朝晖·梦溪村", Latitude: 30.29700, Longitude: 120.16497},
	{ID: "zh-mengxi-7", Campus: CampusZhaohui, Name: "梦溪村7号楼", Category: CategoryDorm, Address: "朝晖·梦溪村", Latitude: 30.29707, Longitude: 120.16497},
	{ID: "zh-zonghe", Campus: CampusZhaohui, Name: "综合楼(学生公寓)", Category: CategoryDorm, Address: "朝晖·生活区", Latitude: 30.29760, Longitude: 120.16530},
	{ID: "mg-library", Campus: CampusMoganshan, Name: "图书馆", Category: CategoryLibrary, Address: "莫干山·中心区", Latitude: 30.54530, Longitude: 119.97100},
	{ID: "mg-dexinfu", Campus: CampusMoganshan, Name: "德馨府餐厅", Category: CategoryDining, Address: "莫干山·生活区", Latitude: 30.54500, Longitude: 119.97140},
	{ID: "mg-gym-main", Campus: CampusMoganshan, Name: "体育馆(主场馆)", Category: CategorySport, Address: "莫干山·文体区", Latitude: 30.54580, Longitude: 119.97120},
	{ID: "mg-gym-sub", Campus: CampusMoganshan, Name: "体育馆(副馆)", Category: CategorySport, Address: "莫干山·文体区", Latitude: 30.54590, Longitude: 119.97080},
	{ID: "mg-track", Campus: CampusMoganshan, Name: "田径场", Category: CategorySport, Address: "莫干山·文体区", Latitude: 30.54610, Longitude: 119.97150},
	{ID: "mg-hospital", Campus: CampusMoganshan, Name: "校医院", Category: CategoryService, Address: "莫干山·生活区", Latitude: 30.54460, Longitude: 119.97130},
	{ID: "mg-teach", Campus: CampusMoganshan, Name: "教学楼区(智慧教室)", Category: CategoryTeaching, Address: "莫干山·教学区", Latitude: 30.54540, Longitude: 119.97070},
	{ID: "mg-dexin-1", Campus: CampusMoganshan, Name: "德馨苑·德一楼", Category: CategoryDorm, Address: "莫干山·德馨苑", Latitude: 30.54480, Longitude: 119.97050},
	{ID: "mg-dexin-2", Campus: CampusMoganshan, Name: "德馨苑·德二楼", Category: CategoryDorm, Address: "莫干山·德馨苑", Latitude: 30.54487, Longitude: 119.97050},
	{ID: "mg-dexin-3", Campus: CampusMoganshan, Name: "德馨苑·德三楼", Category: CategoryDorm, Address: "莫干山·德馨苑", Latitude: 30.54494, Longitude: 119.97050},
	{ID: "mg-dexin-4", Campus: CampusMoganshan, Name: "德馨苑·德四楼", Category: CategoryDorm, Address: "莫干山·德馨苑", Latitude: 30.54501, Longitude: 119.97050},
	{ID: "mg-dexin-5", Campus: CampusMoganshan, Name: "德馨苑·德五楼", Category: CategoryDorm, Address: "莫干山·德馨苑", Latitude: 30.54508, Longitude: 119.97050},
	{ID: "mg-dexin-6", Campus: CampusMoganshan, Name: "德馨苑·德六楼", Category: CategoryDorm, Address: "莫干山·德馨苑", Latitude: 30.54480, Longitude: 119.97057},
	{ID: "mg-dexin-7", Campus: CampusMoganshan, Name: "德馨苑·德七楼", Category: CategoryDorm, Address: "莫干山·德馨苑", Latitude: 30.54487, Longitude: 119.97057},
	{ID: "mg-dexin-8", Campus: CampusMoganshan, Name: "德馨苑·德八楼", Category: CategoryDorm, Address: "莫干山·德馨苑", Latitude: 30.54494, Longitude: 119.97057},
	{ID: "mg-dexin-9", Campus: CampusMoganshan, Name: "德馨苑·德九楼", Category: CategoryDorm, Address: "莫干山·德馨苑", Latitude: 30.54501, Longitude: 119.97057},
	{ID: "mg-dexin-10", Campus: CampusMoganshan, Name: "德馨苑·德十楼", Category: CategoryDorm, Address: "莫干山·德馨苑", Latitude: 30.54508, Longitude: 119.97057},
}
