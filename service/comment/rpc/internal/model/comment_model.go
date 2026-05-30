package model

import (
	"time"
)

type Comment struct {
	ID            uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PostID        uint64     `gorm:"column:post_id;not null;index:idx_post_parent" json:"postId"`
	UserID        uint64     `gorm:"column:user_id;not null;index:idx_user_created" json:"userId"`
	ParentID      uint64     `gorm:"column:parent_id;not null;default:0;index:idx_post_parent,idx_parent_created" json:"parentId"`
	ReplyToUserID *uint64    `gorm:"column:reply_to_user_id" json:"replyToUserId"`
	Level         uint8      `gorm:"column:level;not null;default:1" json:"level"`
	Status        uint8      `gorm:"column:status;not null;default:1" json:"status"`
	ReplyCount    int32      `gorm:"column:reply_count;not null;default:0" json:"replyCount"`
	LikeCount     int32      `gorm:"column:like_count;not null;default:0" json:"likeCount"`
	DislikeCount  int32      `gorm:"column:dislike_count;not null;default:0" json:"dislikeCount"`
	Ip            string     `gorm:"column:ip;size:45;not null;default:''" json:"ip"`
	IpLoc         string     `gorm:"column:ip_loc;size:64;not null;default:''" json:"ipLoc"`
	CreatedAt     time.Time  `gorm:"column:created_at;not null;autoCreateTime" json:"createdAt"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;not null;autoUpdateTime" json:"updatedAt"`
	DeletedAt     *time.Time `gorm:"column:deleted_at" json:"deletedAt"`
}

func (Comment) TableName() string {
	return "comment"
}
