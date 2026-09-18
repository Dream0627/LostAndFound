package pagination

import "strconv"

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

// Parse 解析 page 与 page_size,page 至少为 1,page_size 默认 20、上限 100
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
