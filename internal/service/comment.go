// Package service 是“业务逻辑层”。
// 职责：做业务规则校验、权限判断，并把“仓库层的领域错误”翻译成“对外业务错误(apperror)”。
// 它不直接接触 SQL(那是 repository 的事)，也不处理 HTTP(那是 handler 的事)，
// 这样业务规则集中在一处，便于复用与测试。
package service

import (
	"errors"
	"strings"

	"LAF/internal/model"
	"LAF/internal/repository"
	"LAF/pkg/apperror"
)

// CommentService 依赖两个仓库：
//   - repository     ：评论自身的增删查；
//   - postRepository ：校验评论所属帖子是否存在(发评论前必须先确认帖子存在)。
// 通过构造函数注入依赖，而不是在内部 new，方便替换与测试。
type CommentService struct {
	repository     *repository.CommentRepository
	postRepository *repository.PostRepository
}

// NewCommentService 由 router 在装配阶段调用，注入两个仓库。
func NewCommentService(repository *repository.CommentRepository, postRepository *repository.PostRepository) *CommentService {
	return &CommentService{
		repository:     repository,
		postRepository: postRepository,
	}
}

// CreateCommentInput 是“发表评论”的入参 DTO(数据传输对象)。
// 用 DTO 而不是直接收 HTTP 请求体，可以把“接口层字段”和“业务层字段”解耦。
// PostID 表示评论所针对的帖子。
type CreateCommentInput struct {
	PostID  uint64 `json:"post_id"`
	Content string `json:"content"`
}

// Create 发表评论。userID 是当前登录用户(由中间件从令牌解析而来)，不由前端传入，防止冒充他人。
// 业务步骤：
//   1) 校验目标帖子存在(不存在则直接返回帖子不存在的错误)；
//   2) 清洗并校验评论内容(去首尾空格，长度需在 1~1000 之间)；
//   3) 组装评论实体并落库；
//   4) 返回创建好的评论。
func (s *CommentService) Create(userID uint64, input CreateCommentInput) (*model.Comment, error) {
	if _, err := s.postRepository.GetPostByID(input.PostID); err != nil {
		return nil, err
	}

	input.Content = strings.TrimSpace(input.Content) // 去掉首尾空白，避免“只有空格”被当成有效内容
	if len(input.Content) == 0 || len(input.Content) > 1000 {
		return nil, apperror.ParamError
	}

	comment := &model.Comment{
		PostID:  input.PostID,
		UserID:  userID,
		Content: input.Content,
	}

	if err := s.repository.Create(comment); err != nil {
		return nil, err
	}

	return comment, nil
}

// GetCommentByID 按 ID 取评论，并把仓库的“评论不存在”翻译成对外的业务错误。
func (s *CommentService) GetCommentByID(commentID uint64) (*model.Comment, error) {
	comment, err := s.repository.GetCommentByID(commentID)
	if errors.Is(err, repository.ErrCommentNotFound) {
		return nil, apperror.CommentNotFoundError
	}
	if err != nil {
		return nil, err
	}
	return comment, nil
}

// CommentListResult 是评论列表的返回结构(list + 分页信息)，直接序列化给前端。
type CommentListResult struct {
	List     []*model.Comment `json:"list"`
	Total    int64            `json:"total"`
	Page     int              `json:"page"`
	PageSize int              `json:"page_size"`
}

// GetCommentsByPostID 分页查询某帖子的评论。
// 把对外的 page/page_size 换算成数据库需要的 limit/offset：offset = (page-1)*pageSize。
func (s *CommentService) GetCommentsByPostID(postID uint64, page, pageSize int) (*CommentListResult, error) {
	offset := (page - 1) * pageSize
	comments, total, err := s.repository.GetCommentsByPostID(postID, pageSize, offset)
	if err != nil {
		return nil, err
	}
	return &CommentListResult{
		List:     comments,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// CheckCommentPermission 判断当前用户是否有权删除该评论。
// 规则：student 只能删自己的评论；postadmin / mainadmin 可删任意评论；其余一律禁止。
// 无权时返回 UserForbiddenError(403)。
func (s *CommentService) CheckCommentPermission(userID uint64, role string, comment *model.Comment) error {
	if role == "student" {
		if comment.UserID == userID {
			return nil
		}
	} else if role == "postadmin" || role == "mainadmin" {
		return nil
	}
	return apperror.UserForbiddenError
}

// DeleteComment 执行删除(实为软删除，见仓库层)。
func (s *CommentService) DeleteComment(commentID uint64) error {
	return s.repository.DeleteComment(commentID)
}
