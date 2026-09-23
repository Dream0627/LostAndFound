// Package model 定义与数据库表一一对应的数据结构(GORM 模型)。
// 约定：
//   - gorm 标签描述列名、类型、约束、索引与中文列注释；
//   - json 标签控制返回给前端的字段名，标为 "-" 的字段不会出现在 JSON 中(例如密码哈希)；
//   - 每个实体都带 DeletedAt(gorm.DeletedAt) 字段，实现“软删除”：
//     删除只是把 deleted_at 记为当前时间，数据仍保留在库里，可用 Unscoped 查询或恢复。
package model

import (
	"time"

	"gorm.io/gorm"
)

// 帖子审核状态的三种取值。集中定义为常量，避免在代码各处散落硬编码字符串(拼错很难发现)。
// pending 待审核，approved 已通过，rejected 已驳回。
const (
	PostStatusPending  = "pending"
	PostStatusApproved = "approved"
	PostStatusRejected = "rejected"
)

// 申诉原因的三种取值。self_regret：自行注销反悔；wrongful_ban：被管理员误封号、请求撤回；other：其他(需自行填写说明)。
const (
	AppealReasonSelfRegret  = "self_regret"
	AppealReasonWrongfulBan = "wrongful_ban"
	AppealReasonOther       = "other"
)

// 申诉审核状态的三种取值：pending 待审核、approved 已通过、rejected 已驳回。
const (
	AppealStatusPending  = "pending"
	AppealStatusApproved = "approved"
	AppealStatusRejected = "rejected"
)

// User 对应用户表。角色 Role 取值：student(学生)、postadmin(帖子管理员)、mainadmin(超级管理员)。
// 密码以哈希形式保存(PasswordHash)，绝不明文存储；其 json 标签为 "-"，保证哈希不会返回给前端。
type User struct {
	ID           uint64         `gorm:"primaryKey;autoIncrement;comment:用户ID" json:"id"`
	Username     string         `gorm:"column:username;type:varchar(32);unique;not null,comment:学号或管理员工号" json:"username"`
	Name         string         `gorm:"column:name;type:varchar(32);not null;size:32" json:"name"`
	PasswordHash string         `gorm:"column:password_hash;type:varchar(255);not null;comment:密码哈希值" json:"-"` // json:"-" 表示序列化时忽略该字段，避免密码哈希泄露
	Role         string         `gorm:"column:role;type:enum('student','postadmin','mainadmin');not null;default:'student';comment:角色" json:"role"`
	CreatedAt    time.Time      `gorm:"column:created_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);comment:创建时间" json:"created_at"`
	UpdatedAt    time.Time      `gorm:"column:updated_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);onUpdate:CURRENT_TIMESTAMP(3);comment:更新时间" json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}

// Post 对应帖子表(失物/招领)。
// Type 区分 lost(寻物)与 found(招领)；Status 是审核状态(见上面的常量)；
// ImageURL 使用指针 *string，因为图片是可选的——NULL 表示“没有图片”，与空字符串语义不同。
type Post struct {
	ID         uint64         `gorm:"primaryKey;autoIncrement;comment:帖子ID" json:"id"`
	UserID     uint64         `gorm:"column:user_id;not null;index;comment:作者ID" json:"user_id"`
	Type       string         `gorm:"column:type;type:enum('lost','found');not null;default:'lost';comment:帖子类型" json:"type"`
	Title      string         `gorm:"column:title;type:varchar(2000);not null" json:"title"`
	ImageURL   *string        `gorm:"column:image_url;type:varchar(1024);default:null;comment:图片URL" json:"image_url"` // 用指针类型表示“可为空”，NULL 与空字符串含义不同
	LocationID string `gorm:"column:location_id;type:varchar(64);not null;default:'';comment:校园预设地点ID(冗余快照)" json:"location_id"`
	LocationName string `gorm:"column:location_name;type:varchar(128);not null;default:'';comment:地点名称快照(冗余,减少前端查询)" json:"location_name"`
	Supplement string `gorm:"column:supplement;type:varchar(200);not null;default:'';comment:地点补充说明" json:"supplement"`
	Content    string         `gorm:"column:content;type:varchar(2000);not null" json:"content"`
	IsFinished bool           `gorm:"column:is_finished;type:bool;default:false" json:"is_finished"`
	Status     string         `gorm:"column:status;type:enum('pending','approved','rejected');not null;default:'pending';index;comment:审核状态" json:"status"`
	CreatedAt  time.Time      `gorm:"column:created_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);index" json:"created_at"`
	UpdatedAt  time.Time      `gorm:"column:updated_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);onUpdate:CURRENT_TIMESTAMP(3);index" json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}

// Comment 对应评论表。PostID 关联所属帖子，UserID 关联评论作者。
// 没有外键约束(靠应用层保证一致性)，字段上建了索引以加速按帖子查询评论。
type Comment struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement;comment:评论ID" json:"id"`
	PostID    uint64         `gorm:"column:post_id;not null;index;comment:所属帖子ID" json:"post_id"`
	UserID    uint64         `gorm:"column:user_id;not null;index;comment:评论作者ID" json:"user_id"`
	Content   string         `gorm:"column:content;type:varchar(1000);not null;comment:评论内容" json:"content"`
	CreatedAt time.Time      `gorm:"column:created_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);onUpdate:CURRENT_TIMESTAMP(3);comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}

// Appeal 对应申诉表。用户账号被注销(软删除)后无法登录，因此申诉由被注销者以 username 公开提交；
// 超级管理员审核通过后，系统会据 UserID 自动恢复该账号。
// Reason 取值见上方常量；Content 是申诉说明；Status 是审核状态(见上方常量)。
type Appeal struct {
	ID        uint64         `gorm:"primaryKey;autoIncrement;comment:申诉ID" json:"id"`
	UserID    uint64         `gorm:"column:user_id;not null;index;comment:申诉人ID" json:"user_id"`
	Reason    string         `gorm:"column:reason;type:enum('self_regret','wrongful_ban','other');not null;default:'other';comment:申诉原因" json:"reason"`
	Content   string         `gorm:"column:content;type:varchar(1000);not null;default:'';comment:申诉说明" json:"content"`
	Status    string         `gorm:"column:status;type:enum('pending','approved','rejected');not null;default:'pending';index;comment:审核状态" json:"status"`
	CreatedAt time.Time      `gorm:"column:created_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);comment:创建时间" json:"created_at"`
	UpdatedAt time.Time      `gorm:"column:updated_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);onUpdate:CURRENT_TIMESTAMP(3);comment:更新时间" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}

// 完成寻找申请的三种状态：pending 待处理、agreed 已同意、rejected 已拒绝。
const (
	FinishRequestStatusPending  = "pending"
	FinishRequestStatusAgreed   = "agreed"
	FinishRequestStatusRejected = "rejected"
)

// Conversation 对应“申领/召领对话”表。
// 一条对话由某用户在某个帖子下发起的“申领(found)/召领(lost)”开启，
// 连接发起方(InitiatorID)与帖子作者(OwnerID，冗余存储便于按人查询)。
// 同一用户对同一帖子只会有一条对话(应用层幂等保证)。
type Conversation struct {
	ID          uint64         `gorm:"primaryKey;autoIncrement;comment:对话ID" json:"id"`
	PostID      uint64         `gorm:"column:post_id;not null;index;comment:所属帖子ID" json:"post_id"`
	InitiatorID uint64         `gorm:"column:initiator_id;not null;index;comment:发起方(申领/召领人)ID" json:"initiator_id"`
	OwnerID     uint64         `gorm:"column:owner_id;not null;index;comment:帖子作者ID" json:"owner_id"`
	CreatedAt   time.Time      `gorm:"column:created_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);comment:创建时间" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"column:updated_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);onUpdate:CURRENT_TIMESTAMP(3);comment:更新时间" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}

// Message 对应对话消息表。
// ConversationID 关联所属对话，SenderID 是发送者；Content 为消息正文。
type Message struct {
	ID             uint64         `gorm:"primaryKey;autoIncrement;comment:消息ID" json:"id"`
	ConversationID uint64         `gorm:"column:conversation_id;not null;index;comment:所属对话ID" json:"conversation_id"`
	SenderID       uint64         `gorm:"column:sender_id;not null;index;comment:发送者ID" json:"sender_id"`
	Content        string         `gorm:"column:content;type:varchar(1000);not null;comment:消息内容" json:"content"`
	CreatedAt      time.Time      `gorm:"column:created_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);comment:创建时间" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);onUpdate:CURRENT_TIMESTAMP(3);comment:更新时间" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}

// FinishRequest 对应“完成寻找申请”表。
// 对话任一方可发起；另一方处理(agreed/rejected)。同意后对应帖子被置为已完成(is_finished=true)。
type FinishRequest struct {
	ID             uint64         `gorm:"primaryKey;autoIncrement;comment:完成申请ID" json:"id"`
	ConversationID uint64         `gorm:"column:conversation_id;not null;index;comment:所属对话ID" json:"conversation_id"`
	RequesterID    uint64         `gorm:"column:requester_id;not null;index;comment:发起人ID" json:"requester_id"`
	Status         string         `gorm:"column:status;type:enum('pending','agreed','rejected');not null;default:'pending';index;comment:申请状态" json:"status"`
	CreatedAt      time.Time      `gorm:"column:created_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);comment:创建时间" json:"created_at"`
	UpdatedAt      time.Time      `gorm:"column:updated_at;type:datetime(3);not null;default:CURRENT_TIMESTAMP(3);onUpdate:CURRENT_TIMESTAMP(3);comment:更新时间" json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"column:deleted_at;index" json:"deleted_at"`
}
