// 本文件是公告的数据访问层，包含增删查改以及“软删除/恢复”“审核状态更新”等操作。
// 同样遵循：把数据库错误翻译成领域错误(apperror)再向上抛出。
package repository

import (
	//"errors"
	//"fmt"

	//"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"LAF/internal/model"
	"LAF/pkg/apperror"
)

type AnnouncementRepository struct {
	db *gorm.DB
}

// NewAnnouncementRepository 是构造函数，由 router 注入 db。
func NewAnnouncementRepository(db *gorm.DB) *AnnouncementRepository {
	return &AnnouncementRepository{db: db}
}
func (r *AnnouncementRepository) Create(announcement *model.Announcement) error {
	if err := r.db.Create(announcement).Error; err != nil {
		return apperror.DatabaseError
	}
	return nil
}

func (r *AnnouncementRepository) Delete(announcementID uint64) error {
	if err := r.db.Delete(&model.Announcement{}, announcementID).Error; err != nil {
		return apperror.DatabaseError
	}
	return nil
}
func (r *AnnouncementRepository) GetAnnouncementsByID(announcementsID uint64) (*model.Announcement, error) {
	var announcement model.Announcement
	if err := r.db.First(&announcement, announcementsID).Error; err != nil {
		return nil, apperror.AnnouncementNotFoundError
	}
	if !announcement.DeletedAt.Valid {
		return nil, apperror.AnnouncementNotDeletedError
	}
	return &announcement, nil
}
func (r *AnnouncementRepository) GetAnnouncements(limit, offset int) ([]*model.Announcement, int64, error) {
	var Announcements []*model.Announcement
	var total int64

	if err := r.db.Model(&model.Announcement{}).Count(&total).Error; err != nil {
		return nil, 0, apperror.DatabaseError
	}
	if err := r.db.Model(&model.Announcement{}).Order("created_at desc, id desc").Limit(limit).Offset(offset).Find(&Announcements).Error; err != nil {
		return nil, 0, apperror.DatabaseError
	}
	r.fillAnnouncementAuthorNames(Announcements) // 回填作者姓名(展示用，非表字段)
	return Announcements, total, nil
}
