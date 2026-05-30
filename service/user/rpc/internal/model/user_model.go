package model

import (
	"gorm.io/gorm"
	"time"
)

// User 用户基础表
type User struct {
	ID            uint64         `gorm:"column:id;primaryKey;autoIncrement;comment:用户ID"`
	Username      string         `gorm:"column:username;uniqueIndex:uk_username;size:32;not null;default:'';comment:用户名（唯一）"`
	Phone         *string        `gorm:"column:phone;uniqueIndex:uk_phone;size:16;comment:手机号"` // 可空字段用指针
	Email         *string        `gorm:"column:email;uniqueIndex:uk_email;size:64;comment:邮箱"`  // 可空字段用指针
	PasswordHash  string         `gorm:"column:password_hash;size:255;not null;default:'';comment:加密密码（bcrypt）"`
	Avatar        string         `gorm:"column:avatar;size:255;not null;default:'';comment:头像URL"`
	Status        int8           `gorm:"column:status;default:1;index:idx_status;comment:状态：1正常，2禁言，3冻结"`
	IsAdmin       int8           `gorm:"column:is_admin;default:0;comment:是否管理员：0否，1是"`
	LastLoginIP   *string        `gorm:"column:last_login_ip;size:45;default:'';comment:最后登录IP"` // 可空字段用指针
	LastLoginTime *time.Time     `gorm:"column:last_login_time;comment:最后登录时间戳"`                 // 可空字段用指针
	CreatedAt     time.Time      `gorm:"column:created_at;autoCreateTime;comment:注册时间"`
	UpdatedAt     time.Time      `gorm:"column:updated_at;autoUpdateTime;comment:更新时间"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index;comment:删除时间（软删除）"`
}

// TableName 指定表名
func (User) TableName() string {
	return "user"
}
