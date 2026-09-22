// 本文件是用户的数据访问层：注册写入、按用户名/ID 查询、更新资料等。
package repository

import (
	"errors"
	"time"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"

	"LAF/internal/model"
	"LAF/pkg/apperror"
)

// 两个哨兵错误：用户不存在、用户已存在。定义成固定值便于上层用 errors.Is 判断。
var (
	ErrUserNotFound   = errors.New("user not found")
	ErrUserExists     = errors.New("user already exists")
	ErrUserNotDeleted = errors.New("user is not deactivated")
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
	var mysqlErr *mysql.MySQLError
	if errors.Is(err, gorm.ErrDuplicatedKey) || (errors.As(err, &mysqlErr) && mysqlErr.Number == 1062) {
		return ErrUserExists
	}
	if err != nil {
		return apperror.DatabaseError
	}
	return nil
}

// FindByUsernameUnscoped 按用户名查询用户，且“包含已软删除”的账号。
// 申诉接口要针对“已被注销(软删除)的账号”，普通查询会被 GORM 自动过滤掉 deleted_at 非空的行，
// 因此这里用 Unscoped() 跳过默认过滤，才能查到被注销的账号。查不到时返回 ErrUserNotFound。
func (r *UserRepository) FindByUsernameUnscoped(username string) (*model.User, error) {
	var user model.User
	err := r.db.Unscoped().Where("username = ?", username).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, apperror.DatabaseError
	}
	return &user, nil
}

// GetUserByIDUnscoped 按主键查询用户，且“包含已软删除”的账号。
// 恢复用户时需要读取其 deleted_at(作为“本批删除时间戳”)，故必须能看到已删除的记录。
func (r *UserRepository) GetUserByIDUnscoped(userID uint64) (*model.User, error) {
	var user model.User
	err := r.db.Unscoped().Where("id = ?", userID).First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ErrUserNotFound
	}
	if err != nil {
		return nil, apperror.DatabaseError
	}
	return &user, nil
}

// DeactivateUser 注销用户：在同一事务里，把“用户本人、其发布的帖子、其发表的评论”
// 三者的 deleted_at 一并置为同一个时间戳(now)。
// 为什么用同一个时间戳？这样在恢复时可以用 deleted_at = now 精确圈定“本次注销同批删除的内容”，
// 而不会误恢复该用户此前因别的原因单独删除的帖子/评论。
// 为什么放在事务里？三步更新必须“要么都成功、要么都回滚”，否则会出现“用户被注销但帖子还在”的半删状态。
// 注意：只处理 user_id 属于该用户的帖子与评论；他人写在“该用户帖子下”的评论不在此范围内(保持可见)。
func (r *UserRepository) DeactivateUser(userID uint64) error {
	now := time.Now().Truncate(time.Millisecond) // 截断到毫秒，与数据库 datetime(3) 精度一致，保证恢复时能精确匹配
	err := r.db.Transaction(func(tx *gorm.DB) error {
		// Unscoped() 跳过 GORM 对软删除的自动过滤；显式只更新目标行/目标集合。
		if err := tx.Unscoped().Model(&model.User{}).Where("id = ?", userID).Update("deleted_at", now).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Model(&model.Post{}).Where("user_id = ? AND deleted_at IS NULL", userID).Update("deleted_at", now).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Model(&model.Comment{}).Where("user_id = ? AND deleted_at IS NULL", userID).Update("deleted_at", now).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return apperror.DatabaseError
	}
	return nil
}

// RecoverUser 恢复用户：把用户自身的 deleted_at 置回 NULL，
// 并“只恢复与该用户同批删除(deleted_at 等于用户原 deleted_at)的帖子与评论”。
// 关键点：先从数据库读回用户原始的 deleted_at 作为本批时间戳，再据此恢复内容，
// 这样既不依赖调用方传入时间，也避免 Go 时间戳与库中毫秒值因精度差异而匹配失败。
// 若用户不存在返回 ErrUserNotFound；若用户本就未被注销(无可恢复)返回 ErrUserNotDeleted。
func (r *UserRepository) RecoverUser(userID uint64) error {
	var user model.User
	if err := r.db.Unscoped().Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return apperror.DatabaseError
	}
	if !user.DeletedAt.Valid {
		return ErrUserNotDeleted // 用户当前是正常状态，无需恢复
	}
	batchTime := user.DeletedAt.Time // 用户原删除时间 = 本次注销批次的时间戳

	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Unscoped().Model(&model.User{}).Where("id = ?", userID).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		// deleted_at = batchTime：精确命中“与用户同批删除”的内容，避免误恢复其它时间删除的帖子/评论。
		if err := tx.Unscoped().Model(&model.Post{}).Where("user_id = ? AND deleted_at = ?", userID, batchTime).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		if err := tx.Unscoped().Model(&model.Comment{}).Where("user_id = ? AND deleted_at = ?", userID, batchTime).Update("deleted_at", nil).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return apperror.DatabaseError
	}
	return nil
}
