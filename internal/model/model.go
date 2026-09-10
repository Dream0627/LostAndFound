package model

import "time"

type User struct {
	ID           uint64   `gorm:"primaryKey;autoIncrement;comment:用户ID" json:"id"`
	Username     string `gorm:"column:username;type:varchar(32);unique;not null,comment:学号或管理员工号" json:"username"`
	Name         string `gorm:"column:name;type:varchar(32);not null;size:32" json:"name"`
	PasswordHash string `gorm:"column:password_hash;type:varchar(255);not null;comment:密码哈希值" json:"-"`
	Role         string `gorm:"column:role;type:enum('student','postadmin','mainadmin');not null;default:'student';comment:角色" json:"role"`
	CreatedAt    time.Time  `gorm:"column:created_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"column:updated_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);onUpdate:CURRENT_TIMESTAMP(3);comment:更新时间" json:"updated_at"`
}

type Post struct {
    ID        uint64    `gorm:"primaryKey;autoIncrement;comment:帖子ID" json:"id"`
    UserID    uint      `gorm:"column:user_id;not null;index;comment:作者ID" json:"user_id"`
    Type      string    `gorm:"column:type;type:enum('lost','found');not null;default:'lost';comment:帖子类型" json:"type"` 
	Title     string    `gorm:"column:title;type:varchar(2000);not null" json:"title"`
    ImageURL  *string   `gorm:"column:image_url;type:varchar(1024);default:null;comment:图片URL" json:"image_url"`       
    Content   string    `gorm:"column:content;type:varchar(2000);not null" json:"content"`
    CreatedAt time.Time `gorm:"column:created_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);index" json:"created_at"`
    UpdatedAt time.Time `gorm:"column:updated_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);onUpdate:CURRENT_TIMESTAMP(3);index" json:"updated_at"`
}