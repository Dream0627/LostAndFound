// 本文件是帖子的业务逻辑层：创建、查询、可见性控制、权限判断、删除与恢复。
package service

import (
	//"errors"
	//"fmt"
	
	"strings"
	//"time"
	"LAF/internal/model"
	"LAF/internal/repository"
	"LAF/pkg/apperror"
)

// PostService 只依赖帖子仓库(评论相关的联动删除在仓库层事务里完成)。
type PostService struct { 
	repository *repository.PostRepository
}

// NewPostService 由 router 注入帖子仓库。
func NewPostService(repository *repository.PostRepository) *PostService { 
	return &PostService{repository: repository}
}

// GetPostByID 直连仓库查询未删除帖子，供删除/恢复前做权限判断使用。
func (s *PostService) GetPostByID(postID uint64) (*model.Post, error) { 
	return s.repository.GetPostByID(postID)
}

// GetPostByIDUnscoped 查询“含已软删除”的帖子，用于恢复场景。
func (s *PostService) GetPostByIDUnscoped(postID uint64) (*model.Post, error) { 
	return s.repository.GetPostByIDUnscoped(postID)
}

// CheckpostPermission 判断当前用户是否有权操作(删除/恢复)该帖子。
// 规则：student 只能操作自己的帖子；postadmin / mainadmin 可操作任意帖子。
func (s *PostService) CheckpostPermission(userID uint64, role string, post *model.Post) error { 
	if role == "student" { 
		if post.UserID == userID { 
			return nil
		}
	} else if role == "postadmin" || role == "mainadmin" { 
		return nil
	}
    return apperror.UserForbiddenError
}

// CreateInput 是“发布帖子”的业务入参 DTO。ImageURL 用指针表示“可选(可为空)”。
type CreateInput struct {
    Type     string
    Title    string
    Content  string
    ImageURL *string
}

// Create 发布帖子。核心业务规则：
//   1) type 只能是 lost / found；
//   2) 内容去空白后长度需在 1~2000；
//   3) 审核状态：普通学生发布默认 pending(待审核)，
//      管理员发布直接 approved(免审核)。
func (s *PostService) Create(input CreateInput, userID uint64, role string) (*model.Post, error) {
	if input.Type != "lost" && input.Type != "found" {
		return nil, apperror.ParamError
	}
	input.Content = strings.TrimSpace(input.Content)
	if len(input.Content) == 0 || len(input.Content) > 2000 {
		return nil, apperror.ParamError
	}

	status := model.PostStatusPending // 默认待审核；管理员另外处理
	if isPostAdmin(role) {
		status = model.PostStatusApproved
	}

	post := &model.Post{
		Type:       input.Type,
		Title:      input.Title,
		ImageURL:   input.ImageURL,
		Content:    input.Content,
		IsFinished: false,
		Status:     status,
		UserID:     userID,
	}

	if err := s.repository.Create(post); err != nil {
		return nil, err
	}

	return post, nil
}

// isPostAdmin 判断角色是否为管理员(postadmin 或 mainadmin)。抽成小函数便于复用与统一口径。
func isPostAdmin(role string) bool {
	return role == "postadmin" || role == "mainadmin"
}

// PostListResult 是帖子列表的返回结构(list + 分页信息)。
type PostListResult struct {
	List     []*model.Post `json:"list"`
	Total    int64         `json:"total"`
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
}

// GetPosts 分页查询帖子，并在业务层做“可见性/过滤”控制：
//   - type 过滤：只接受 lost / found，非法值报参数错误；
//   - status 过滤：管理员可指定任意合法状态；普通用户被强制只看 approved，
//     从而保证未审核/被驳回的帖子不会泄漏给普通用户。
//   - 若最终没有任何状态条件，则默认放开全部状态(主要用于管理员场景)。
func (s *PostService) GetPosts(types []string, statuses []string, role string, page, pageSize int) (*PostListResult, error) {
	validTypes := make([]string, 0, len(types)) // 逐个校验 type，剔除/拦截非法值
	for _, postType := range types {
		if postType != "lost" && postType != "found" {
			return nil, apperror.ParamError
		}
		validTypes = append(validTypes, postType)
	}

	validStatuses := make([]string, 0, len(statuses)) // 按角色决定可见的状态集合
	if isPostAdmin(role) {
		for _, status := range statuses {
			if status != model.PostStatusPending && status != model.PostStatusApproved && status != model.PostStatusRejected {
				return nil, apperror.InvalidPostStatusError
			}
			validStatuses = append(validStatuses, status)
		}
	} else {
		validStatuses = append(validStatuses, model.PostStatusApproved)
	}

	if len(validStatuses) == 0 {
		validStatuses = append(validStatuses, model.PostStatusPending, model.PostStatusApproved, model.PostStatusRejected)
	}

	offset := (page - 1) * pageSize
	posts, total, err := s.repository.GetPosts(validTypes, validStatuses, pageSize, offset)
	if err != nil {
		return nil, err
	}

	return &PostListResult{
		List:     posts,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// GetVisiblePost 查询“对当前用户可见”的帖子详情。
// 关键点：如果调用者不是管理员，且帖子状态不是 approved，就当作“不存在”返回。
// 用“不存在”而非“无权限”，可避免暴露“这个 id 上其实有个未审核帖子”的信息。
func (s *PostService) GetVisiblePost(postID uint64, role string) (*model.Post, error) {
	post, err := s.repository.GetPostByID(postID)
	if err != nil {
		return nil, err
	}
	if !isPostAdmin(role) && post.Status != model.PostStatusApproved {
		return nil, apperror.PostNotFoundError
	}
	return post, nil
}

// DeletePost 委托仓库删除(含级联软删评论)。
func (s *PostService) DeletePost(postID uint64) error { 
	return s.repository.DeletePost(postID)
}

// RecoverPost 委托仓库恢复被软删除的帖子。
func (s *PostService) RecoverPost(postID uint64) error { 
	return s.repository.RecoverPost(postID)
}