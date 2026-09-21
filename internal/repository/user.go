// 本文件是用户的数据访问层：注册写入、按用户名/ID 查询、更新资料等。
package repository

import (
	"errors"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"LAF/internal/model"
	"LAF/pkg/apperror"
)

// 两个哨兵错误：用户不存在、用户已存在。定义成固定值便于上层用 errors.Is 判断。
var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExists   = errors.New("user already exists")
)

// UserRepository 持有数据库句柄 db，为用户提供数据访问能力。
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository 是构造函数，由 router 注入 db。
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create 插入用户。这里特意捕获“唯一键冲突(用户名重复)”：
// 既兼容 GORM 的 ErrDuplicatedKey，也兼容 MySQL 原生错误号 1062，
// 统一翻译成领域错误 ErrUserExists，避免把数据库细节泄漏到上层。
func (r *UserRepository) Create(user *model.User) error {
	err := r.db.Create(user).Error
	var mysqlErr *mysql.MySQLError
	if errors.Is(err, gorm.ErrDuplicatedKey) || errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return ErrUserExists
	}
	if err != nil {
		return apperror.DatabaseError
	}
	return nil
}

// FindByUsername 按用户名查询用户(登录、注册查重都会用到)。
// 查不到时返回 ErrUserNotFound，供上层区分“没找到”和其他错误。
func (r *UserRepository) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, apperror.DatabaseError
	}
	return &user, nil
}

// GetPostsByUserID 查询某用户发布的所有帖子，按 id 倒序(新的在前)。
func (r *UserRepository) GetPostsByUserID(userID uint64) ([]*model.Post, error) { 
	var posts []*model.Post
	err := r.db.Order("id desc").Where("user_id = ?", userID).Find(&posts).Error
	if err != nil { 
		return nil, apperror.DatabaseError
	}
	return posts, nil
}

// GetProfile 按主键查询用户(用于个人资料页)。
func (r *UserRepository) GetProfile(userID uint64) (*model.User, error) { 
	var user model.User
	err := r.db.Where("id = ?", userID).First(&user).Error
	if err != nil { 
		return nil, apperror.DatabaseError
	}
	return &user, nil
}

// GetUserByID 按主键查询用户(用于改密等需要读取原密码哈希的场景)。
func (r *UserRepository) GetUserByID(userID uint64) (*model.User, error) { 
	var user model.User
	err := r.db.Where("id = ?", userID).First(&user).Error
	if err != nil { 
		return nil, apperror.DatabaseError
	}
	return &user, nil
}

// UpdateProfile 做“部分更新”：只更新传入 map 里出现的字段。
// 用 map[string]interface{} 而不是整个结构体，是为了避免把未填字段覆盖成零值。
func (r *UserRepository) UpdateProfile(userID uint64, updates map[string]interface{}) error { 
	err := r.db.Model(&model.User{}).Where("id = ?", userID).Updates(updates).Error
	if err != nil { 
		return apperror.DatabaseError
	}
	return nil
}