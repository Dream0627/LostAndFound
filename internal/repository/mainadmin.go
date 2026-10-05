// 预留文件：超级管理员相关的数据访问。
// 目前暂无实现，保留占位以维持目录结构。
package repository

import (
	"LAF/internal/model"

	"gorm.io/gorm"
)

type StatsRepository struct {
	db *gorm.DB
}

func NewStatsRepository(db *gorm.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

type Stats struct {
	UserCount          int64
	PostCount          int64
	PendingPostCount   int64
	PendingAppealCount int64
	TodayPostCount     int64
	TodayCommentCount  int64
}

func (r *StatsRepository) GetStats() (stat Stats, err error) {
	var userCount, postCount, todayPostCount, todayCommentCount, pendingPostCount, pendingAppealCount int64
	if err = r.db.Model(&model.User{}).Where("deleted_at IS NULL").Count(&userCount).Error; err != nil {
		return Stats{}, err
	}
	if err = r.db.Model(&model.Post{}).Where("deleted_at IS NULL").Count(&postCount).Error; err != nil {
		return Stats{}, err
	}
	if err = r.db.Model(&model.Post{}).Where("deleted_at IS NULL AND status = ? ", model.PostStatusPending).Count(&pendingPostCount).Error; err != nil {
		return Stats{}, err
	}
	if err = r.db.Model(&model.Appeal{}).Where("deleted_at IS NULL AND status = ? ", model.AppealStatusPending).Count(&pendingAppealCount).Error; err != nil {
		return Stats{}, err
	}
	if err = r.db.Model(&model.Post{}).
		Where("DATE(created_at) = CURDATE() AND deleted_at IS NULL").
		Count(&todayPostCount).Error; err != nil {
		return Stats{}, err
	}
	if err = r.db.Model(&model.Comment{}).
		Where("DATE(created_at) = CURDATE() AND deleted_at IS NULL").
		Count(&todayCommentCount).Error; err != nil {
		return Stats{}, err
	}
	return Stats{
		UserCount:          userCount,
		PostCount:          postCount,
		PendingPostCount:   pendingPostCount,
		PendingAppealCount: pendingAppealCount,
		TodayPostCount:     todayPostCount,
		TodayCommentCount:  todayCommentCount,
	}, nil
}
