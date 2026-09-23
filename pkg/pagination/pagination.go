// Package pagination 统一解析列表接口的分页参数。
// 对外约定：page 从 1 开始，page_size 默认 20、上限 100。
// 统一规则的好处是各列表接口分页行为一致，同时用上限防止一次拉取过多数据拖垮数据库。
package pagination

import "strconv"

// 分页的默认值(20)与上限(100)。上限用于防止恶意或误传的超大 page_size。
const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Parse 把 URL 查询参数里的字符串解析成分页整数。
// 解析失败或取值不合法时一律回退到安全默认值（不报错），使接口更健壮：
//   page 非法或小于 1      -> 1
//   page_size 非法或小于 1 -> 默认 20
//   page_size 超过上限     -> 截断为 100
func Parse(pageValue, pageSizeValue string) (int, int) {
	page := 1
	pageSize := DefaultPageSize

	if n, err := strconv.Atoi(pageValue); err == nil && n > 0 {
		page = n
	}
	if n, err := strconv.Atoi(pageSizeValue); err == nil && n > 0 {
		pageSize = n
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	return page, pageSize
}

// Offset 把对外的 page/page_size 换算成数据库查询所需的偏移量：offset = (page-1)*pageSize。
// 集中成一个函数可避免各列表服务各写一遍同样的算式、口径不一。
// page 小于 1 时按第 1 页处理(page/page_size 已由 Parse 归一，这里再兜底一次)。
func Offset(page, pageSize int) int {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		return 0
	}
	return (page - 1) * pageSize
}
