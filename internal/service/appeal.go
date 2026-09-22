// 本文件是申诉的业务逻辑层：提交申诉(公开接口)与相关校验。
package service

import (
	"errors"
	"strings"

	"LAF/internal/model"
	"LAF/internal/repository"
	"LAF/pkg/apperror"
)

// AppealService 依赖两个仓库：
//   - repository     ：申诉自身的增删查；
//   - userRepository ：根据 username 找到申诉指向的账号(须能看到已注销的账号)。
type AppealService struct {
	repository     *repository.AppealRepository
	userRepository *repository.UserRepository
}

// NewAppealService 由 router 在装配阶段调用，注入两个仓库。
func NewAppealService(repository *repository.AppealRepository, userRepository *repository.UserRepository) *AppealService {
	return &AppealService{
		repository:     repository,
		userRepository: userRepository,
	}
}

// CreateAppealInput 是提交申诉的入参 DTO。之所以用 username 而不是 userID：
// 账号被注销后用户无法登录、也拿不到自己的 ID，只能用学号/工号来指明账号。
type CreateAppealInput struct {
	Username string
	Reason   string
	Content  string
}

// isValidAppealReason 判断申诉原因是否为三种合法取值之一。
func isValidAppealReason(reason string) bool {
	return reason == model.AppealReasonSelfRegret ||
		reason == model.AppealReasonWrongfulBan ||
		reason == model.AppealReasonOther
}

// Create 提交申诉。业务规则：
//  1. username 去空白后长度需在 1~32；
//  2. reason 必须是三种合法取值之一；
//  3. content 去空白后长度需 ≤1000；当 reason=other(其他)时必须自行填写说明，不得为空；
//  4. username 对应的账号必须存在，且必须处于“已注销(软删除)”状态——申诉只针对被注销的账号；
//     因此这里用 Unscoped 查询(能看到已删除账号)，若账号仍处于正常状态则拒绝申诉。
func (s *AppealService) Create(input CreateAppealInput) (*model.Appeal, error) {
	input.Username = strings.TrimSpace(input.Username)
	if len(input.Username) == 0 || len(input.Username) > 32 {
		return nil, apperror.ParamError
	}

	if !isValidAppealReason(input.Reason) {
		return nil, apperror.InvalidAppealReasonError
	}

	input.Content = strings.TrimSpace(input.Content)
	if len(input.Content) > 1000 {
		return nil, apperror.ParamError
	}
	if input.Reason == model.AppealReasonOther && len(input.Content) == 0 {
		return nil, apperror.ParamError // “其他”原因必须写明具体缘由
	}

	// 用 Unscoped 查询，才能查到“已被注销(软删除)”的账号。
	user, err := s.userRepository.FindByUsernameUnscoped(input.Username)
	if errors.Is(err, repository.ErrUserNotFound) {
		return nil, apperror.UserNotFoundError
	}
	if err != nil {
		return nil, err
	}

	// 只有被注销的账号才可申诉；正常账号没有可申诉的对象。
	if !user.DeletedAt.Valid {
		return nil, apperror.UserNotDeactivatedError
	}

	appeal := &model.Appeal{
		UserID:  user.ID,
		Reason:  input.Reason,
		Content: input.Content,
		Status:  model.AppealStatusPending, // 新提交的申诉默认待审核
	}

	if err := s.repository.Create(appeal); err != nil {
		return nil, err
	}
	return appeal, nil
}
