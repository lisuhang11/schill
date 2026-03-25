package model

import (
	"time"
)

type Topic struct {
	ID        uint64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name      string     `gorm:"column:name;size:255;not null;uniqueIndex:uk_name" json:"name"`
	QuoteNum  int64      `gorm:"column:quote_num;not null;default:0;index:idx_quote_num" json:"quoteNum"`
	CreatedAt time.Time  `gorm:"column:created_at;not null;autoCreateTime" json:"createdAt"`
	UpdatedAt time.Time  `gorm:"column:updated_at;not null;autoUpdateTime" json:"updatedAt"`
	DeletedAt *time.Time `gorm:"column:deleted_at" json:"deletedAt"`
}

func (Topic) TableName() string {
	return "topic"
}
