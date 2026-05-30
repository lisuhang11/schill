package model

import (
	"time"
)

type CommentVote struct {
	ID        uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	CommentID uint64     `gorm:"column:comment_id;not null;index:idx_comment_user,idx_user_comment" json:"commentId"`
	UserID    uint64     `gorm:"column:user_id;not null;index:idx_comment_user,idx_user_comment" json:"userId"`
	VoteType  uint8      `gorm:"column:vote_type;not null" json:"voteType"`
	CreatedAt time.Time  `gorm:"column:created_at;not null;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null;autoUpdateTime" json:"updatedAt"`
}

func (CommentVote) TableName() string {
	return "comment_vote"
}
