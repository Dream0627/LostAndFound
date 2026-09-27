// 本文件是“申领/召领对话 + 完成寻找”的业务逻辑层。
// 设计要点：申领(found 帖子)与召领(lost 帖子)共用同一套“发起对话”逻辑，
// 由帖子类型自动判定，无需在接口层区分。
package service

import (
	"strings"

	"LAF/internal/model"
	"LAF/internal/repository"
	"LAF/pkg/apperror"
	"LAF/pkg/pagination"
)

// ConversationService 依赖对话/消息/完成申请三个仓库，以及帖子仓库(用于校验帖子与置为已完成)。
type ConversationService struct {
	conversationRepository  *repository.ConversationRepository
	messageRepository       *repository.MessageRepository
	finishRequestRepository *repository.FinishRequestRepository
	postRepository          *repository.PostRepository
}

// NewConversationService 由 router 注入各仓库。
func NewConversationService(
	conversationRepository *repository.ConversationRepository,
	messageRepository *repository.MessageRepository,
	finishRequestRepository *repository.FinishRequestRepository,
	postRepository *repository.PostRepository,
) *ConversationService {
	return &ConversationService{
		conversationRepository:  conversationRepository,
		messageRepository:       messageRepository,
		finishRequestRepository: finishRequestRepository,
		postRepository:          postRepository,
	}
}

// StartConversation 对指定帖子发起申领/召领，开启与楼主的对话。
// 规则：
//  1. 帖子必须可见(普通用户仅见 approved，管理员可见任意)；
//  2. 帖子必须未完成(is_finished = false)，已完成则拒绝；
//  3. 不能对自己发布的帖子发起；
//  4. 幂等：同一用户对同一帖子重复发起，直接返回已存在的对话。
func (s *ConversationService) StartConversation(postID, userID uint64, role string) (*model.Conversation, error) {
	post, err := s.postRepository.GetPostByID(postID)
	if err != nil {
		return nil, err
	}
	if err := ensurePostVisible(post, role); err != nil {
		return nil, err
	}
	if post.IsFinished {
		return nil, apperror.PostAlreadyFinishedError
	}
	if post.UserID == userID {
		return nil, apperror.PostClaimSelfError
	}

	existing, err := s.conversationRepository.GetConversationByPostAndInitiator(postID, userID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return existing, nil // 幂等：已发起过，直接复用
	}

	conversation := &model.Conversation{
		PostID:      postID,
		InitiatorID: userID,
		OwnerID:     post.UserID,
	}
	if err := s.conversationRepository.Create(conversation); err != nil {
		return nil, err
	}
	return conversation, nil
}

// ConversationListResult 是对话列表的返回结构(list + 分页信息)。
// ConversationListResult 是对话列表的返回结构，复用通用分页结构 PageResult。
type ConversationListResult = PageResult[*model.Conversation]

// GetMyConversations 分页查询当前用户参与(发起或收到)的对话。
func (s *ConversationService) GetMyConversations(userID uint64, page, pageSize int) (*ConversationListResult, error) {
	offset := pagination.Offset(page, pageSize)
	conversations, total, err := s.conversationRepository.ListConversationsByUser(userID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	for _, conversation := range conversations {
		s.fillPostSnapshot(conversation) // 回填帖子标题/完成状态，供会话列表展示
	}
	return &ConversationListResult{List: conversations, Total: total, Page: page, PageSize: pageSize}, nil
}

// GetConversationDetail 查询单个对话(仅参与方)，并回填所属帖子快照。
// 供聊天页展示帖子标题、跳转原帖、在帖子已完成时隐藏“发起完成寻找”入口。
func (s *ConversationService) GetConversationDetail(conversationID, userID uint64) (*model.Conversation, error) {
	conversation, err := s.conversationRepository.GetConversationByID(conversationID)
	if err != nil {
		return nil, err
	}
	if err := s.checkParticipant(conversation, userID); err != nil {
		return nil, err
	}
	s.fillPostSnapshot(conversation)
	return conversation, nil
}

// fillPostSnapshot 把帖子标题/审核状态/完成状态回填到对话的非表字段。
// 帖子可能已被删除，此时查询报错则保留零值，不影响对话本身的访问。
func (s *ConversationService) fillPostSnapshot(conversation *model.Conversation) {
	if conversation == nil {
		return
	}
	post, err := s.postRepository.GetPostByID(conversation.PostID)
	if err != nil {
		return
	}
	conversation.PostTitle = post.Title
	conversation.PostStatus = post.Status
	conversation.PostFinished = post.IsFinished
}

// addSystemMessage 往对话里写一条系统消息(SenderID 为 NULL 表示系统)。
// 系统消息仅用于过程留痕，写入失败不应阻断主流程，故忽略错误。
func (s *ConversationService) addSystemMessage(conversationID uint64, content string) {
	_ = s.messageRepository.Create(&model.Message{
		ConversationID: conversationID,
		SenderID:       nil,
		Content:        content,
	})
}

// checkParticipant 校验用户是否为该对话的参与方(发起方或楼主)，否则返回无权访问。
func (s *ConversationService) checkParticipant(conversation *model.Conversation, userID uint64) error {
	if conversation.InitiatorID != userID && conversation.OwnerID != userID {
		return apperror.ConversationForbiddenError
	}
	return nil
}

// MessageListResult 是消息列表的返回结构(list + 分页信息)。
// MessageListResult 是消息列表的返回结构，复用通用分页结构 PageResult。
type MessageListResult = PageResult[*model.Message]

// GetMessages 分页查询某对话的消息(仅参与方可读)。
func (s *ConversationService) GetMessages(conversationID, userID uint64, page, pageSize int) (*MessageListResult, error) {
	conversation, err := s.conversationRepository.GetConversationByID(conversationID)
	if err != nil {
		return nil, err
	}
	if err := s.checkParticipant(conversation, userID); err != nil {
		return nil, err
	}

	offset := pagination.Offset(page, pageSize)
	messages, total, err := s.messageRepository.ListMessagesByConversation(conversationID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	return &MessageListResult{List: messages, Total: total, Page: page, PageSize: pageSize}, nil
}

// SendMessage 在对话中发送一条消息(仅参与方可发)。
func (s *ConversationService) SendMessage(conversationID, userID uint64, content string) (*model.Message, error) {
	conversation, err := s.conversationRepository.GetConversationByID(conversationID)
	if err != nil {
		return nil, err
	}
	if err := s.checkParticipant(conversation, userID); err != nil {
		return nil, err
	}

	content = strings.TrimSpace(content)
	if len(content) == 0 || len(content) > 1000 {
		return nil, apperror.ParamError
	}

	message := &model.Message{
		ConversationID: conversationID,
		SenderID:       &userID,
		Content:        content,
	}
	if err := s.messageRepository.Create(message); err != nil {
		return nil, err
	}
	return message, nil
}

// CreateFinishRequest 在对话中发起“完成寻找”申请(任一方均可发起)。
// 规则：帖子尚未完成；同一对话不得存在待处理的申请。
func (s *ConversationService) CreateFinishRequest(conversationID, userID uint64) (*model.FinishRequest, error) {
	conversation, err := s.conversationRepository.GetConversationByID(conversationID)
	if err != nil {
		return nil, err
	}
	if err := s.checkParticipant(conversation, userID); err != nil {
		return nil, err
	}

	post, err := s.postRepository.GetPostByID(conversation.PostID)
	if err != nil {
		return nil, err
	}
	if post.IsFinished {
		return nil, apperror.PostAlreadyFinishedError
	}

	pending, err := s.finishRequestRepository.GetPendingFinishRequestByConversation(conversationID)
	if err != nil {
		return nil, err
	}
	if pending != nil {
		return nil, apperror.FinishRequestNotPendingError // 已有待处理申请，避免重复发起
	}

	request := &model.FinishRequest{
		ConversationID: conversationID,
		RequesterID:    userID,
		Status:         model.FinishRequestStatusPending,
	}
	if err := s.finishRequestRepository.Create(request); err != nil {
		return nil, err
	}
	// 过程留痕：在对话中生成一条系统消息，双方都能看到。
	s.addSystemMessage(conversationID, "一方发起了“完成寻找”申请，等待对方确认")
	return request, nil
}

// ReviewFinishRequest 由对话中“另一方”处理完成申请：同意(agreed)则置帖子为已完成，拒绝(rejected)则不改动帖子。
// 规则：只有非发起方能处理；申请必须仍为 pending。
func (s *ConversationService) ReviewFinishRequest(conversationID, requestID, userID uint64, status string) (*model.FinishRequest, error) {
	if status != model.FinishRequestStatusAgreed && status != model.FinishRequestStatusRejected {
		return nil, apperror.InvalidFinishRequestStatusError
	}

	conversation, err := s.conversationRepository.GetConversationByID(conversationID)
	if err != nil {
		return nil, err
	}
	if err := s.checkParticipant(conversation, userID); err != nil {
		return nil, err
	}

	request, err := s.finishRequestRepository.GetFinishRequestByID(requestID)
	if err != nil {
		return nil, err
	}
	if request.ConversationID != conversationID {
		return nil, apperror.FinishRequestNotFoundError
	}
	if request.Status != model.FinishRequestStatusPending {
		return nil, apperror.FinishRequestNotPendingError
	}
	if request.RequesterID == userID {
		return nil, apperror.ConversationForbiddenError // 发起方不能处理自己的申请
	}

	if err := s.finishRequestRepository.UpdateStatus(requestID, status); err != nil {
		return nil, err
	}
	request.Status = status

	if status == model.FinishRequestStatusAgreed {
		if err := s.postRepository.SetPostFinished(conversation.PostID, true); err != nil {
			return nil, err
		}
		s.addSystemMessage(conversationID, "“完成寻找”申请已通过，帖子已标记为已完成")
	} else {
		s.addSystemMessage(conversationID, "“完成寻找”申请被拒绝，双方可继续沟通后重新发起")
	}
	return request, nil
}

// GetPendingFinishRequest 查询某对话当前待处理的完成申请(仅参与方可见)。
// 不存在待处理申请时返回 (nil, nil)，供前端进入会话时判断是否显示申请横幅。
func (s *ConversationService) GetPendingFinishRequest(conversationID, userID uint64) (*model.FinishRequest, error) {
	conversation, err := s.conversationRepository.GetConversationByID(conversationID)
	if err != nil {
		return nil, err
	}
	if err := s.checkParticipant(conversation, userID); err != nil {
		return nil, err
	}
	return s.finishRequestRepository.GetPendingFinishRequestByConversation(conversationID)
}

// WithdrawFinishRequest 由申请发起方撤回自己发起的“完成寻找”申请。
// 规则：只有发起方能撤回；申请必须仍为 pending；撤回后申请被软删除，帖子不受影响。
func (s *ConversationService) WithdrawFinishRequest(conversationID, requestID, userID uint64) error {
	conversation, err := s.conversationRepository.GetConversationByID(conversationID)
	if err != nil {
		return err
	}
	if err := s.checkParticipant(conversation, userID); err != nil {
		return err
	}

	request, err := s.finishRequestRepository.GetFinishRequestByID(requestID)
	if err != nil {
		return err
	}
	if request.ConversationID != conversationID {
		return apperror.FinishRequestNotFoundError
	}
	if request.Status != model.FinishRequestStatusPending {
		return apperror.FinishRequestNotPendingError
	}
	if request.RequesterID != userID {
		return apperror.UserForbiddenError // 只有发起方能撤回自己的申请
	}
	if err := s.finishRequestRepository.Delete(requestID); err != nil {
		return err
	}
	s.addSystemMessage(conversationID, "发起方已撤回“完成寻找”申请")
	return nil
}