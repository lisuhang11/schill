package model

import (
	"time"
)

type Post struct {
	ID              uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID          uint64     `gorm:"column:user_id;not null;index:idx_user_id" json:"userId"`
	CommentCount    uint64     `gorm:"column:comment_count;not null;default:0" json:"commentCount"`
	CollectionCount uint64     `gorm:"column:collection_count;not null;default:0" json:"collectionCount"`
	UpvoteCount     uint64     `gorm:"column:upvote_count;not null;default:0" json:"upvoteCount"`
	ShareCount      uint64     `gorm:"column:share_count;not null;default:0" json:"shareCount"`
	Visibility      int32      `gorm:"column:visibility;not null;default:0" json:"visibility"`
	IsTop           int32      `gorm:"column:is_top;not null;default:0" json:"isTop"`
	IsEssence       int32      `gorm:"column:is_essence;not null;default:0" json:"isEssence"`
	IsLock          int32      `gorm:"column:is_lock;not null;default:0" json:"isLock"`
	LatestRepliedAt int64      `gorm:"column:latest_replied_at;not null;default:0" json:"latestRepliedAt"`
	Tags            string     `gorm:"column:tags;size:255;not null;default:''" json:"tags"`
	Ip              string     `gorm:"column:ip;size:45;not null;default:''" json:"ip"`
	IpLoc           string     `gorm:"column:ip_loc;size:64;not null;default:''" json:"ipLoc"`
	CreatedAt       time.Time  `gorm:"column:created_at;not null;autoCreateTime" json:"createdAt"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;not null;autoUpdateTime" json:"updatedAt"`
	DeletedAt       *time.Time `gorm:"column:deleted_at" json:"deletedAt"`
}

func (Post) TableName() string {
	return "post"
}
