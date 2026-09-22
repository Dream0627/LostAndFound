// 本文件是用户相关的业务逻辑：注册、登录(签发 JWT)、资料查询与更新、修改密码。
package service

import (
	"errors"
	//"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"LAF/config"
	"LAF/internal/model"
	"LAF/internal/repository"
	"LAF/pkg/apperror"
)

// 预先编译好的正则：要求用户名必须“全是数字”(本项目学号/工号均为数字)。
// 预编译一次可避免每次注册都重新编译，提升效率。
var numericUsername = regexp.MustCompile(`^[0-9]+$`)

// UserService 依赖用户仓库和 JWT 配置(登录时要用密钥与有效期签发令牌)。
type UserService struct {
	repository *repository.UserRepository
	jwt        config.JWTConfig
}

// RegisterInput 是注册的业务入参 DTO。
type RegisterInput struct {
	Username string
	Name     string
	Password string
	Role     string
}

// LoginResult 是登录成功后的返回：令牌本身、令牌类型、有效期与用户信息。
type LoginResult struct {
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
	ExpiresIn   int64       `json:"expires_in"`
	User        *model.User `json:"user"`
}

// tokenClaims 是签发的 JWT 载荷：user_id、role 加上标准声明。
type tokenClaims struct {
	UserID uint64 `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// NewUserService 由 router 注入用户仓库与 JWT 配置。
func NewUserService(repository *repository.UserRepository, jwtConfig config.JWTConfig) *UserService {
	return &UserService{
		repository: repository,
		jwt:        jwtConfig,
	}
}

// Register 处理注册。业务规则：
//   - 用户名/姓名去空白后长度需在 1~32；
//   - 密码长度需在 8~16；
//   - 用户名必须是纯数字；
//   - 用户名不能重复(先查一次，插入时再兜底捕获唯一键冲突)；
//   - 密码必须经 bcrypt 哈希后再入库，绝不存明文。
//
// 注意：当前“仅允许注册 student 角色”的校验被注释掉了(临时放开)。
func (s *UserService) Register(input RegisterInput) (*model.User, error) {
	input.Username = strings.TrimSpace(input.Username)
	input.Name = strings.TrimSpace(input.Name)
	if len(input.Username) == 0 || len(input.Username) > 32 {
		return nil, apperror.ParamError
	}
	if len(input.Name) == 0 || len(input.Name) > 32 {
		return nil, apperror.ParamError
	}
	if len(input.Password) < 8 || len(input.Password) > 16 {
		return nil, apperror.ParamError
	}
	if !numericUsername.MatchString(input.Username) {
		return nil, apperror.ParamError
	}
	// if input.Role != "student" {
	// 	return nil, apperror.ParamError
	// }

	if _, err := s.repository.FindByUsername(input.Username); err == nil {
		return nil, apperror.UserRepeatError
	} else if !errors.Is(err, repository.ErrUserNotFound) {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	user := &model.User{
		Username:     input.Username,
		Name:         input.Name,
		PasswordHash: string(passwordHash),
		Role:         input.Role,
	}
	if err := s.repository.Create(user); err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			return nil, apperror.UserRepeatError
		}
		return nil, err
	}
	return user, nil
}

// Login 校验账号密码并签发 JWT。
// 说明：用户不存在与密码错误返回同一个 LoginError，避免被用来探测“某用户名是否存在”。
// 校验通过后用 HS256 签发令牌，载荷含 user_id 与 role，并设置签发时间与过期时间。
func (s *UserService) Login(username, password string) (*LoginResult, error) {
	user, err := s.repository.FindByUsername(strings.TrimSpace(username))
	if errors.Is(err, repository.ErrUserNotFound) {
		return nil, apperror.LoginError
	}
	if err != nil {
		return nil, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return nil, apperror.LoginError
	}
	if s.jwt.Secret == "" || s.jwt.ExpireSeconds <= 0 {
		return nil, errors.New("invalid jwt configuration")
	}

	now := time.Now()
	claims := tokenClaims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.Username,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(s.jwt.ExpireSeconds) * time.Second)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(s.jwt.Secret))
	if err != nil {
		return nil, err
	}
	return &LoginResult{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   s.jwt.ExpireSeconds,
		User:        user,
	}, nil
}

// GetPostsByUserID 查询某用户发布的帖子，供个人资料页展示。
func (s *UserService) GetPostsByUserID(userID uint64) ([]*model.Post, error) {
	posts, err := s.repository.GetPostsByUserID(userID)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

// GetProfileResult 是个人资料页的返回：用户信息 + 其发布的帖子。
type GetProfileResult struct {
	User  *model.User   `json:"user"`
	Posts []*model.Post `json:"posts"`
}

// GetProfile 组合“用户信息”与“其帖子列表”一并返回。
func (s *UserService) GetProfile(userID uint64) (*GetProfileResult, error) {
	user, err := s.repository.GetProfile(userID)
	if err != nil {
		return nil, err
	}

	userPosts, err := s.GetPostsByUserID(userID)
	if err != nil {
		return nil, err
	}

	data := &GetProfileResult{User: user, Posts: userPosts}

	return data, nil
}

// GetUserByID 按 ID 查询用户，供改密等需要读取原密码哈希的逻辑使用。
func (s *UserService) GetUserByID(userID uint64) (*model.User, error) {
	return s.repository.GetUserByID(userID)
}

// UpdateProfileInput 是更新资料的入参：只允许改姓名与用户名。
type UpdateProfileInput struct {
	Name     string `json:"name"`
	Username string `json:"username"`
}

// UpdateProfile 只更新“本次传了值”的字段：
// 用 map 收集非空字段，空字符串表示“不改该项”。
// 若一个字段都没传，则直接返回，不发无意义的数据库更新。
func (s *UserService) UpdateProfile(userID uint64, input UpdateProfileInput) error {

	updates := make(map[string]interface{})

	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.Username != "" {
		updates["username"] = input.Username
	}

	if len(updates) == 0 {
		return nil
	}

	err := s.repository.UpdateProfile(userID, updates)
	if errors.Is(err, repository.ErrUserExists) {
		return apperror.UserRepeatError
	}
	return err
}

// UpdatePasswordInput 是修改密码的入参：原密码、新密码、确认密码。
type UpdatePasswordInput struct {
	OldPassword     string `json:"old_password"`
	NewPassword     string `json:"new_password"`
	ConfirmPassword string `json:"confirm_password"`
}

// UpdatePassword 修改密码。业务规则：
//  1. 新密码长度需在 8~16；
//  2. 两次输入的新密码必须一致；
//  3. 必须校验原密码正确，防止会话被劫持后直接改密；
//  4. 新密码同样经 bcrypt 哈希后入库。
func (s *UserService) UpdatePassword(userID uint64, input UpdatePasswordInput) error {
	if len(input.NewPassword) < 8 || len(input.NewPassword) > 16 {
		return apperror.ParamError
	}

	if input.NewPassword != input.ConfirmPassword {
		return apperror.ParamError
	}

	user, err := s.repository.GetUserByID(userID)
	if errors.Is(err, repository.ErrUserNotFound) {
		return apperror.UserNotFoundError
	}
	if err != nil {
		return err
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.OldPassword)) != nil {
		return apperror.OldPasswordError
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repository.UpdateProfile(userID, map[string]interface{}{
		"password_hash": string(passwordHash),
	})
}

// DeactivateAccount 注销本人账号(软删除)。业务规则：
//  1. 账号必须存在且当前处于正常状态(未注销)，否则返回相应错误；
//  2. 注销会“同批软删除本人账号及其名下帖子、评论”(详见仓库层 DeactivateUser)。
//
// 注销后该账号无法登录；若想恢复，用户需提交申诉，由超级管理员审核通过后级联恢复。
func (s *UserService) DeactivateAccount(userID uint64) error {
	user, err := s.repository.GetUserByIDUnscoped(userID)
	if errors.Is(err, repository.ErrUserNotFound) {
		return apperror.UserNotFoundError
	}
	if err != nil {
		return err
	}
	if user.DeletedAt.Valid {
		return apperror.UserAlreadyDeactivatedError
	}
	return s.repository.DeactivateUser(userID)
}
