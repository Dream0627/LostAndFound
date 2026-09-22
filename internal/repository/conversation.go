// 本文件是“申领/召领对话”的数据访问层。
// 遵循项目约定：把数据库错误翻译成领域错误(apperror)再向上抛出。
package repository

import (
	"errors"

	"gorm.io/gorm"

	"LAF/internal/model"
	"LAF/pkg/apperror"
)

// ConversationRepository 持有数据库句柄 db，为对话提供数据访问能力。
type ConversationRepository struct {
	db *gorm.DB
}

// NewConversationRepository 是构造函数，由 router 注入 db。
func NewConversationRepository(db *gorm.DB) *ConversationRepository {
	return &ConversationRepository{db: db}
}

// GetConversationByID 查询未删除的对话，查不到返回 apperror.ConversationNotFoundError。
func (r *ConversationRepository) GetConversationByID(conversationID uint64) (*model.Conversation, error) {
	var conversation model.Conversation
	err := r.db.Where("id = ?", conversationID).First(&conversation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.ConversationNotFoundError
		}
		return nil, apperror.DatabaseError
	}
	return &conversation, nil
}

// GetConversationByPostAndInitiator 查询“某用户对某帖子”已建立的对话，用于幂等创建。
// 不存在时返回 (nil, nil)，交由上层决定是否新建。
func (r *ConversationRepository) GetConversationByPostAndInitiator(postID, initiatorID uint64) (*model.Conversation, error) {
	var conversation model.Conversation
	err := r.db.Where("post_id = ? AND initiator_id = ?", postID, initiatorID).First(&conversation).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperror.DatabaseError
	}
	return &conversation, nil
}

// Create 插入一条对话，自增主键会回填到 conversation.ID。
func (r *ConversationRepository) Create(conversation *model.Conversation) error {
	if err := r.db.Create(conversation).Error; err != nil {
		return apperror.DatabaseError
	}
	return nil
}

// ListConversationsByUser 分页查询“我参与的对话”(作为发起方或楼主)，按 id 倒序(新对话在前)。
func (r *ConversationRepository) ListConversationsByUser(userID uint64, limit, offset int) ([]*model.Conversation, int64, error) {
	conversations := make([]*model.Conversation, 0)
	var total int64

	buildQuery := func() *gorm.DB {
		return r.db.Model(&model.Conversation{}).Where("initiator_id = ? OR owner_id = ?", userID, userID)
	}

	if err := buildQuery().Count(&total).Error; err != nil {
		return nil, 0, apperror.DatabaseError
	}
	if err := buildQuery().Order("id desc").Limit(limit).Offset(offset).Find(&conversations).Error; err != nil {
		return nil, 0, apperror.DatabaseError
	}
	return conversations, total, nil
}
