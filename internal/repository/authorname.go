package repository

// 本文件集中处理“把 user_id 解析成作者姓名”的逻辑。
// 帖子表/评论表只存 user_id，不建外键、也不做 ORM 预加载(Preload/Joins)；
// 这里在应用层一次性批量查用户表，把姓名回填到模型的 AuthorName 字段
// (该字段标注 gorm:"-"，不落库)，供接口直接返回给前端展示。

import (
	"gorm.io/gorm"

	"LAF/internal/model"
)

// authorNameUnknown 作者查不到时的展示占位(例如作者账号已被软删除)。
const authorNameUnknown = "未知用户"

// collectUserIDs 对用户ID去重，减少一次查询里的重复条件。
func collectUserIDs(raw []uint64) []uint64 {
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

// loadUserNames 按用户ID集合批量查询姓名，返回 id -> name 映射。
// 查询出错或用户缺失都不抛异常：调用方会用占位名兜底，避免影响帖子/评论主流程。
func loadUserNames(db *gorm.DB, ids []uint64) map[uint64]string {
	names := make(map[uint64]string, len(ids))
	if len(ids) == 0 {
		return names
	}
	var users []*model.User
	if err := db.Model(&model.User{}).Where("id IN ?", ids).Find(&users).Error; err != nil {
		return names
	}
	for _, u := range users {
		names[u.ID] = u.Name
	}
	return names
}

// fillPostAuthorNames 为一批帖子批量填充作者姓名(AuthorName)。
func (r *PostRepository) fillPostAuthorNames(posts []*model.Post) {
	if len(posts) == 0 {
		return
	}
	ids := make([]uint64, 0, len(posts))
	for _, p := range posts {
		if p != nil {
			ids = append(ids, p.UserID)
		}
	}
	names := loadUserNames(r.db, collectUserIDs(ids))
	for _, p := range posts {
		if p == nil {
			continue
		}
		if name, ok := names[p.UserID]; ok && name != "" {
			p.AuthorName = name
		} else {
			p.AuthorName = authorNameUnknown
		}
	}
}

// fillCommentAuthorNames 为一批评论批量填充作者姓名(AuthorName)。
func (r *CommentRepository) fillCommentAuthorNames(comments []*model.Comment) {
	if len(comments) == 0 {
		return
	}
	ids := make([]uint64, 0, len(comments))
	for _, c := range comments {
		if c != nil {
			ids = append(ids, c.UserID)
		}
	}
	names := loadUserNames(r.db, collectUserIDs(ids))
	for _, c := range comments {
		if c == nil {
			continue
		}
		if name, ok := names[c.UserID]; ok && name != "" {
			c.AuthorName = name
		} else {
			c.AuthorName = authorNameUnknown
		}
	}
}
