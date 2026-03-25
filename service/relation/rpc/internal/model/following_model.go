package model

import (
	"time"
)

type Following struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    uint64    `gorm:"column:user_id;not null;uniqueIndex:uk_user_follow" json:"userId"`
	FollowID  uint64    `gorm:"column:follow_id;not null;uniqueIndex:uk_user_follow;index:idx_follow" json:"followId"`
	CreatedAt time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time `gorm:"column:updated_at;not null;autoUpdateTime" json:"updatedAt"`
}

func (Following) TableName() string {
	return "following"
}
