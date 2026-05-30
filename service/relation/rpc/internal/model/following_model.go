package model

import (
	"time"
)

type Following struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"column:user_id;not null;uniqueIndex:uk_user_follow" json:"userId"`
	FollowID  uint64    `gorm:"column:follow_id;not null;uniqueIndex:uk_user_follow;index:idx_follow" json:"followId"`
	IsMutual  bool      `gorm:"column:is_mutual;not null;default:false" json:"isMutual"`
	CreatedAt time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"createdAt"`
}

func (Following) TableName() string {
	return "following"
}
