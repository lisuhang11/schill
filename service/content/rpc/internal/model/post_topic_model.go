package model

import (
	"time"
)

type PostTopic struct {
	ID        uint64    `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	PostID    uint64    `gorm:"column:post_id;not null;index:idx_post" json:"postId"`
	TopicID   uint64    `gorm:"column:topic_id;not null;index:idx_topic" json:"topicId"`
	CreatedAt time.Time `gorm:"column:created_at;not null;autoCreateTime" json:"createdAt"`
}

func (PostTopic) TableName() string {
	return "post_topic"
}
