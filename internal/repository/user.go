package repository

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"LAF/internal/model"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(user *model.User) error {
	err := r.db.Create(user).Error
	var mysqlErr *mysql.MySQLError
	if errors.Is(err, gorm.ErrDuplicatedKey) || errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrUserExists
	}
	return err
}

func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetPostsByUserID(userID uint64) ([]*model.Post, error) { 
	var posts []*model.Post
	err := r.db.Order("id desc").Where("user_id = ?", userID).Find(&posts).Error
	if err != nil { 
		return nil, err
	}
	return posts, nil
}

func (r *UserRepository) GetProfile(userID uint64) (*model.User, error) { 
	var user model.User
	err := r.db.Where("id = ?", userID).First(&user).Error
	if err != nil { 
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) GetUserByID(userID uint64) (*model.User, error) { 
	var user model.User
	err := r.db.Where("id = ?", userID).First(&user).Error
	if err != nil { 
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateProfile(userID uint64, updates map[string]interface{}) error { 
	err := r.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
	if err != nil { 
		return err
	}
	return nil
}