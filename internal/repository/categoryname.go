package repository

import (
	"LAF/internal/model"

	"gorm.io/gorm"
)

// fillPostCategoryName 为一批帖子批量填充分类名称(CategoryName)。
// 先收集所有非空的 category_id，去重后批量查询分类表，得到 id → name 映射；
// 再遍历每条帖子，把对应的分类名称回填到 CategoryName 字段(该字段标注 gorm:"-"，不落库)，供接口直接返回给前端展示。
func collectCategoryIDs(raw []uint64) []uint64 {
	seen := make(map[uint64]struct{}, len(raw))
	result := make([]uint64, 0, len(raw))
	for _, id := range raw {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

// loadCategoryNames 按分类ID集合批量查询名称，返回 id -> name 映射。
// 查询出错或分类缺失都不抛异常：调用方会用占位名兜底，避免影响帖子/评论主流程。
func loadCategoryNames(db *gorm.DB, ids []uint64) map[uint64]string {
	names := make(map[uint64]string, len(ids))
	if len(ids) == 0 {
		return names
	}
	var categories []*model.Category
	if err := db.Model(&model.Category{}).Where("id IN ?", ids).Find(&categories).Error; err != nil {
		return names
	}
	for _, c := range categories {
		names[c.ID] = c.Name
	}
	return names
}

// fillCategoryNames 为一批帖子批量填充分类名称(CategoryName)。
func (r *PostRepository) fillCategoryNames(posts []*model.Post) {
	if len(posts) == 0 {
		return
	}
	ids := make([]uint64, 0, len(posts))
	for _, p := range posts {
		if p != nil {
			ids = append(ids, p.CategoryID)
		}
	}
	names := loadCategoryNames(r.db, collectCategoryIDs(ids))
	for _, p := range posts {
		if p == nil {
			continue
		}
		if name, ok := names[p.CategoryID]; ok && name != "" {
			p.CategoryName = name
		} else {
			p.CategoryName = "未分类"
		}
	}
}
