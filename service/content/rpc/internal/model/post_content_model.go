package model

import (
	"time"
)

type PostContent struct {
	ID        uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PostID    uint64     `gorm:"column:post_id;not null;index:idx_post_id" json:"postId"`
	UserID    uint64     `gorm:"column:user_id;not null;index:idx_user_id" json:"userId"`
	Content   string     `gorm:"column:content;type:text;not null" json:"content"`
	Type      int32      `gorm:"column:type;not null;default:2" json:"type"`
	Sort      int32      `gorm:"column:sort;not null;default:100" json:"sort"`
	CreatedAt time.Time  `gorm:"column:created_at;not null;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null;autoUpdateTime" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deletedAt"`
}

func (PostContent) TableName() string {
	return "post_content"
}
