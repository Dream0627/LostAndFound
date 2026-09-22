// 本文件是对话消息的数据访问层。
// 遵循项目约定：把数据库错误翻译成领域错误(apperror)再向上抛出。
package repository

import (
	"errors"

	"gorm.io/gorm"

	"LAF/internal/model"
	"LAF/pkg/apperror"
)

// MessageRepository 持有数据库句柄 db，为对话消息提供数据访问能力。
type MessageRepository struct {
	db *gorm.DB
}

// NewMessageRepository 是构造函数，由 router 注入 db。
func NewMessageRepository(db *gorm.DB) *MessageRepository {
	return &MessageRepository{db: db}
}

// GetMessageByID 查询未删除的消息，查不到返回 apperror.MessageNotFoundError。
func (r *MessageRepository) GetMessageByID(messageID uint64) (*model.Message, error) {
	var message model.Message
	err := r.db.Where("id = ?", messageID).First(&message).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.MessageNotFoundError
		}
		return nil, apperror.DatabaseError
	}
	return &message, nil
}

// Create 插入一条消息，自增主键会回填到 message.ID。
func (r *MessageRepository) Create(message *model.Message) error {
	if err := r.db.Create(message).Error; err != nil {
		return apperror.DatabaseError
	}
	return nil
}

// ListMessagesByConversation 分页查询某对话的消息，按 id 倒序(新消息在前)。
func (r *MessageRepository) ListMessagesByConversation(conversationID uint64, limit, offset int) ([]*model.Message, int64, error) {
	messages := make([]*model.Message, 0)
	var total int64

	buildQuery := func() *gorm.DB {
		return r.db.Model(&model.Message{}).Where("conversation_id = ?", conversationID)
	}

	if err := buildQuery().Count(&total).Error; err != nil {
		return nil, 0, apperror.DatabaseError
	}
	if err := buildQuery().Order("id desc").Limit(limit).Offset(offset).Find(&messages).Error; err != nil {
		return nil, 0, apperror.DatabaseError
	}
	return messages, total, nil
}
