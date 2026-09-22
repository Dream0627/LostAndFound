// 本文件是申诉的数据访问层：创建申诉、按主键查询、更新审核状态、按状态分页查询。
// 与其它 repository 一致：只和数据库打交道，并把数据库错误翻译成领域错误(apperror / 哨兵错误)向上抛出。
package repository

import (
	"errors"

	"gorm.io/gorm"

	"LAF/internal/model"
	"LAF/pkg/apperror"
)

// ErrAppealNotFound 是本层的“哨兵错误”，表示“申诉不存在”。用固定值定义，便于上层用 errors.Is 判断。
var (
	ErrAppealNotFound = errors.New("appeal not found")
)

// AppealRepository 持有数据库句柄 db，为申诉提供数据访问能力。
type AppealRepository struct {
	db *gorm.DB
}

// NewAppealRepository 是构造函数，由 router 在装配时注入 db。
func NewAppealRepository(db *gorm.DB) *AppealRepository {
	return &AppealRepository{db: db}
}

// Create 插入一条申诉，自增主键会回填到 appeal.ID。
func (r *AppealRepository) Create(appeal *model.Appeal) error {
	if err := r.db.Create(appeal).Error; err != nil {
		return apperror.DatabaseError
	}
	return nil
}

// GetAppealByID 按主键查询一条(未删除的)申诉；查不到时返回 ErrAppealNotFound。
func (r *AppealRepository) GetAppealByID(appealID uint64) (*model.Appeal, error) {
	var appeal model.Appeal
	err := r.db.Where("id = ?", appealID).First(&appeal).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrAppealNotFound
	}
	if err != nil {
		return nil, apperror.DatabaseError
	}
	return &appeal, nil
}

// UpdateAppealStatus 更新申诉的审核状态(如 pending -> approved / rejected)。
func (r *AppealRepository) UpdateAppealStatus(appealID uint64, status string) error {
	if err := r.db.Model(&model.Appeal{}).Where("id = ?", appealID).Update("status", status).Error; err != nil {
		return apperror.DatabaseError
	}
	return nil
}

// GetAppeals 分页查询申诉，可按状态过滤(statuses 为空则不过滤)，按 id 倒序(新申诉在前)。
// 用闭包复用“同一套过滤条件”：先 Count 求总数，再 Limit/Offset 取当页数据。
func (r *AppealRepository) GetAppeals(statuses []string, limit, offset int) ([]*model.Appeal, int64, error) {
	var appeals []*model.Appeal
	var total int64

	buildQuery := func() *gorm.DB {
		query := r.db.Model(&model.Appeal{})
		if len(statuses) > 0 {
			query = query.Where("status IN ?", statuses)
		}
		return query
	}

	if err := buildQuery().Count(&total).Error; err != nil {
		return nil, 0, apperror.DatabaseError
	}
	if err := buildQuery().Order("id desc").Limit(limit).Offset(offset).Find(&appeals).Error; err != nil {
		return nil, 0, apperror.DatabaseError
	}
	return appeals, total, nil
}
