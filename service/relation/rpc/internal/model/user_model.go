package model

import (
	"gorm.io/gorm"
	"time"
)

type User struct {
	ID        uint64         `gorm:"column:id;primaryKey"`
	Username  string         `gorm:"column:username"`
	Avatar    string         `gorm:"column:avatar"`
	CreatedAt time.Time      `gorm:"column:created_at"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

func (User) TableName() string {
	return "user"
}
