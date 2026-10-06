// 本文件提供后台概览数据的聚合计数查询。
package repository

import (
	"LAF/internal/model"

	"gorm.io/gorm"
)

type CountRepository struct {
	db *gorm.DB
}

// NewCountRepository 是构造函数，由 router 注入 db。
func NewCountRepository(db *gorm.DB) *CountRepository {
	return &CountRepository{db: db}
}

// Count 是后台概览计数在仓库层的聚合结果(仅数字，不含展示格式)。
type Count struct {
	UserCount          int64
	PostCount          int64
	PendingPostCount   int64
	PendingAppealCount int64
	TodayPostCount     int64
	TodayCommentCount  int64
}

// GetCount 聚合后台概览计数：用户/帖子总数、待审核帖子数、待处理申诉数、今日新增帖子/评论数。
func (r *CountRepository) GetCount() (count Count, err error) {
	var userCount, postCount, todayPostCount, todayCommentCount, pendingPostCount, pendingAppealCount int64
	if err = r.db.Model(&model.User{}).Where("deleted_at IS NULL").Count(&userCount).Error; err != nil {
		return Count{}, err
	}
	if err = r.db.Model(&model.Post{}).Where("deleted_at IS NULL").Count(&postCount).Error; err != nil {
		return Count{}, err
	}
	if err = r.db.Model(&model.Post{}).Where("deleted_at IS NULL AND status = ?", model.PostStatusPending).Count(&pendingPostCount).Error; err != nil {
		return Count{}, err
	}
	if err = r.db.Model(&model.Appeal{}).Where("deleted_at IS NULL AND status = ?", model.AppealStatusPending).Count(&pendingAppealCount).Error; err != nil {
		return Count{}, err
	}
	if err = r.db.Model(&model.Post{}).
		Where("DATE(created_at) = CURDATE() AND deleted_at IS NULL").
		Count(&todayPostCount).Error; err != nil {
		return Count{}, err
	}
	if err = r.db.Model(&model.Comment{}).
		Where("DATE(created_at) = CURDATE() AND deleted_at IS NULL").
		Count(&todayCommentCount).Error; err != nil {
		return Count{}, err
	}
	return Count{
		UserCount:          userCount,
		PostCount:          postCount,
		PendingPostCount:   pendingPostCount,
		PendingAppealCount: pendingAppealCount,
		TodayPostCount:     todayPostCount,
		TodayCommentCount:  todayCommentCount,
	}, nil
}
