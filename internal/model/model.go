package model

import (
	"time"

	"gorm.io/gorm"
)

const (
	PostStatusPending  = "pending"
	PostStatusApproved = "approved"
	PostStatusRejected = "rejected"
)

type User struct {
	ID           uint64    `gorm:"primaryKey;autoIncrement;comment:用户ID" json:"id"`
	Username     string    `gorm:"column:username;type:varchar(32);unique;not null,comment:学号或管理员工号" json:"username"`
	Name         string    `gorm:"column:name;type:varchar(32);not null;size:32" json:"name"`
	PasswordHash string    `gorm:"column:password_hash;type:varchar(255);not null;comment:密码哈希值" json:"-"`
	Role         string    `gorm:"column:role;type:enum('student','postadmin','mainadmin');not null;default:'student';comment:角色" json:"role"`
	CreatedAt    time.Time `gorm:"column:created_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);onUpdate:CURRENT_TIMESTAMP(3);comment:更新时间" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}

type Post struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement;comment:帖子ID" json:"id"`
	UserID     uint64         `gorm:"column:user_id;not null;index;comment:作者ID" json:"user_id"`
	Type       string         `gorm:"column:type;type:enum('lost','found');not null;default:'lost';comment:帖子类型" json:"type"`
	Title      string         `gorm:"column:title;type:varchar(2000);not null" json:"title"`
	ImageURL   *string        `gorm:"column:image_url;type:varchar(1024);default:null;comment:图片URL" json:"image_url"`
	Content    string         `gorm:"column:content;type:varchar(2000);not null" json:"content"`
	IsFinished bool           `gorm:"column:is_finished;type:bool;default:false" json:"is_finished"`
	Status     string         `gorm:"column:status;type:enum('pending','approved','rejected');not null;default:'pending';index;comment:审核状态" json:"status"`
	CreatedAt  time.Time      `gorm:"column:created_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);index" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);onUpdate:CURRENT_TIMESTAMP(3);index" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}

type Comment struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement;comment:评论ID" json:"id"`
	PostID    uint64         `gorm:"column:post_id;not null;index;comment:所属帖子ID" json:"post_id"`
	UserID    uint64         `gorm:"column:user_id;not null;index;comment:评论作者ID" json:"user_id"`
	Content   string         `gorm:"column:content;type:varchar(1000);not null;comment:评论内容" json:"content"`
	CreatedAt time.Time      `gorm:"column:created_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);onUpdate:CURRENT_TIMESTAMP(3);comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}
