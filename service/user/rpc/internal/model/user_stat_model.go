package model

import "time"

// UserStat 用户统计表
type UserStat struct {
	ID              uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	UserID          uint64    `gorm:"column:user_id;uniqueIndex:uk_user_id;not null;comment:用户ID"`
	PostCount       uint32    `gorm:"column:post_count;type:int unsigned;not null;default:0;comment:发帖数"`
	CommentCount    uint32    `gorm:"column:comment_count;type:int unsigned;not null;default:0;comment:评论数"`
	FollowerCount   uint32    `gorm:"column:follower_count;type:int unsigned;not null;default:0;comment:粉丝数"`
	FollowingCount  uint32    `gorm:"column:following_count;type:int unsigned;not null;default:0;comment:关注数"`
	LikeCount       uint32    `gorm:"column:like_count;type:int unsigned;not null;default:0;comment:获赞总数"`
	CollectionCount uint32    `gorm:"column:collection_count;type:int unsigned;not null;default:0;comment:被收藏总数"`
	LastActiveTime  int64     `gorm:"column:last_active_time;not null;default:0;index:idx_last_active;comment:最后活跃时间"`
	CreatedAt       time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt       time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

// TableName 指定表名
func (UserStat) TableName() string {
	return "user_stat"
}
