// 本文件是公告的数据访问层，提供公告的“新增 / 软删除 / 查询(单条、分页)”能力。
// 同样遵循：把数据库错误翻译成领域错误(apperror)再向上抛出。
package repository

import (
	"errors"

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

// Create 插入一条公告，自增主键会回填到 announcement.ID。
func (r *AnnouncementRepository) Create(announcement *model.Announcement) error {
	if err := r.db.Create(announcement).Error; err != nil {
		return apperror.DatabaseError
	}
	return nil
}

// Delete 软删除公告(GORM 自动把 deleted_at 置为当前时间)。
func (r *AnnouncementRepository) Delete(announcementID uint64) error {
	if err := r.db.Delete(&model.Announcement{}, announcementID).Error; err != nil {
		return apperror.DatabaseError
	}
	return nil
}

// GetAnnouncementByID 查询未删除的公告，用于删除前的存在性校验。
// 查不到时返回 apperror.AnnouncementNotFoundError；其他数据库异常统一返回 apperror.DatabaseError。
func (r *AnnouncementRepository) GetAnnouncementByID(announcementID uint64) (*model.Announcement, error) {
	var announcement model.Announcement
	err := r.db.Where("id = ?", announcementID).First(&announcement).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.AnnouncementNotFoundError
		}
		return nil, apperror.DatabaseError
	}
	return &announcement, nil
}

// GetAnnouncements 分页查询公告，按“新公告在前(created_at desc, id desc)”排序，并回填发布者姓名。
func (r *AnnouncementRepository) GetAnnouncements(limit, offset int) ([]*model.Announcement, int64, error) {
	var announcements []*model.Announcement
	var total int64

	if err := r.db.Model(&model.Announcement{}).Count(&total).Error; err != nil {
		return nil, 0, apperror.DatabaseError
	}
	if err := r.db.Model(&model.Announcement{}).Order("created_at desc, id desc").Limit(limit).Offset(offset).Find(&announcements).Error; err != nil {
		return nil, 0, apperror.DatabaseError
	}
	r.fillAnnouncementAuthorNames(announcements) // 回填作者姓名(展示用，非表字段)
	return announcements, total, nil
}
