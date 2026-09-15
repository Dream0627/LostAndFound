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

var numericUsername = regexp.MustCompile(`^[0-9]+$`)

type UserService struct {
	repository *repository.UserRepository
	jwt        config.JWTConfig
}

type RegisterInput struct {
	Username string
	Name     string
	Password string
	Role     string
}

type LoginResult struct {
	AccessToken string `json:"access_token"`
	TokenType  string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	User 	    *model.User `json:"user"`
}

type tokenClaims struct {
	UserID   uint64  `json:"user_id"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func NewUserService(repository *repository.UserRepository, jwtConfig config.JWTConfig) *UserService {
	return &UserService{
		repository: repository,
		jwt:        jwtConfig,
	}
}

func (s *UserService) Register(input RegisterInput) (*model.User, error) {
	input.Username = strings.TrimSpace(input.Username)
	input.Name = strings.TrimSpace(input.Name)
	if len(input.Username) == 0 || len(input.Username) > 32 {
		return nil, apperror.ParamError
	}
	if  len(input.Name) == 0 || len(input.Name) > 32 {
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
		Username: input.Username,
		Name:     input.Name,
		PasswordHash: string(passwordHash),
		Role:     input.Role,
	}
	if err := s.repository.Create(user); err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			return nil, apperror.UserRepeatError
		}
		return nil, err
	}
	return user, nil
}

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
			Subject:  user.Username,
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
		TokenType: "Bearer",
		ExpiresIn: s.jwt.ExpireSeconds,
		User: user,
	}, nil
}