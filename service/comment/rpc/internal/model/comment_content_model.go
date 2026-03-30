package model

import (
	"time"
)

type CommentContent struct {
	ID           uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CommentID    uint64    `gorm:"column:comment_id;not null;uniqueIndex:uk_comment_id" json:"commentId"`
	Content      string    `gorm:"column:content;type:mediumtext;not null" json:"content"`
	ContentType  uint8     `gorm:"column:content_type;not null;default:1" json:"contentType"`
	CreatedAt    time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"createdAt"`
	UpdatedAt    time.Time `gorm:"column:updated_at;not null;autoUpdateTime" json:"updatedAt"`
}

func (CommentContent) TableName() string {
	return "comment_content"
}
