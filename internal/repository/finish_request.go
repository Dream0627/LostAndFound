// 本文件是“完成寻找申请”的数据访问层。
// 遵循项目约定：把数据库错误翻译成领域错误(apperror)再向上抛出。
package repository

import (
	"errors"

	"gorm.io/gorm"

	"LAF/internal/model"
	"LAF/pkg/apperror"
)

// FinishRequestRepository 持有数据库句柄 db，为完成寻找申请提供数据访问能力。
type FinishRequestRepository struct {
	db *gorm.DB
}

// NewFinishRequestRepository 是构造函数，由 router 注入 db。
func NewFinishRequestRepository(db *gorm.DB) *FinishRequestRepository {
	return &FinishRequestRepository{db: db}
}

// GetFinishRequestByID 查询未删除的完成申请，查不到返回 apperror.FinishRequestNotFoundError。
func (r *FinishRequestRepository) GetFinishRequestByID(requestID uint64) (*model.FinishRequest, error) {
	var request model.FinishRequest
	err := r.db.Where("id = ?", requestID).First(&request).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.FinishRequestNotFoundError
		}
		return nil, apperror.DatabaseError
	}
	return &request, nil
}

// GetPendingFinishRequestByConversation 查询某对话“待处理”的完成申请，用于防止重复发起。
// 不存在时返回 (nil, nil)，交由上层决定是否新建。
func (r *FinishRequestRepository) GetPendingFinishRequestByConversation(conversationID uint64) (*model.FinishRequest, error) {
	var request model.FinishRequest
	err := r.db.Where("conversation_id = ? AND status = ?", conversationID, model.FinishRequestStatusPending).First(&request).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperror.DatabaseError
	}
	return &request, nil
}

// Create 插入一条完成申请，自增主键会回填到 request.ID。
func (r *FinishRequestRepository) Create(request *model.FinishRequest) error {
	if err := r.db.Create(request).Error; err != nil {
		return apperror.DatabaseError
	}
	return nil
}

// UpdateStatus 更新完成申请的状态(如 pending -> agreed / rejected)。
func (r *FinishRequestRepository) UpdateStatus(requestID uint64, status string) error {
	if err := r.db.Model(&model.FinishRequest{}).Where("id = ?", requestID).Update("status", status).Error; err != nil {
		return apperror.DatabaseError
	}
	return nil
}

// Delete 软删除一条完成申请(发起方撤回自己的待处理申请时调用)，只置 deleted_at，不物理删除。
func (r *FinishRequestRepository) Delete(requestID uint64) error {
	if err := r.db.Delete(&model.FinishRequest{}, requestID).Error; err != nil {
		return apperror.DatabaseError
	}
	return nil
}