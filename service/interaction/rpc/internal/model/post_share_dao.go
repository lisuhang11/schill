package model

import (
	"context"
	"gorm.io/gorm"
)

type PostShareDAO struct {
	db *gorm.DB
}

func NewPostShareDAO(db *gorm.DB) *PostShareDAO {
	return &PostShareDAO{db: db}
}

func (dao *PostShareDAO) Create(ctx context.Context, share *PostShare) error {
	return dao.db.WithContext(ctx).Create(share).Error
}

func (dao *PostShareDAO) GetDB() *gorm.DB {
	return dao.db
}
