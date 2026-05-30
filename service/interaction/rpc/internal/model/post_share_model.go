package model

import (
	"time"
)

type PostShare struct {
	ID       uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PostID   uint64    `gorm:"column:post_id;not null;index:idx_post_id,idx_user_post" json:"postId"`
	UserID   uint64    `gorm:"column:user_id;not null;index:idx_user_post" json:"userId"`
	SharedAt time.Time `gorm:"column:shared_at;not null;autoCreateTime" json:"sharedAt"`
}

func (PostShare) TableName() string {
	return "post_share"
}
