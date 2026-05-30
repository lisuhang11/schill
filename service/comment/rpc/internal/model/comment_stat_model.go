package model

import (
	"time"
)

type CommentStat struct {
	CommentID     uint64     `gorm:"column:comment_id;primaryKey" json:"commentId"`
	ReplyCount    uint32     `gorm:"column:reply_count;not null;default:0" json:"replyCount"`
	LikeCount     uint32     `gorm:"column:like_count;not null;default:0" json:"likeCount"`
	DislikeCount  uint32     `gorm:"column:dislike_count;not null;default:0" json:"dislikeCount"`
	UpdatedAt     time.Time  `gorm:"column:updated_at;not null;autoUpdateTime" json:"updatedAt"`
}

func (CommentStat) TableName() string {
	return "comment_stat"
}
