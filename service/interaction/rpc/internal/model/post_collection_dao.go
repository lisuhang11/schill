package model

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostCollectionDAO struct {
	db *gorm.DB
}

func NewPostCollectionDAO(db *gorm.DB) *PostCollectionDAO {
	return &PostCollectionDAO{db: db}
}

func (dao *PostCollectionDAO) Create(ctx context.Context, collection *PostCollection) error {
	return dao.db.WithContext(ctx).Create(collection).Error
}

func (dao *PostCollectionDAO) Delete(ctx context.Context, postId, userId uint64) error {
	return dao.db.WithContext(ctx).
		Where("post_id = ? AND user_id = ?", postId, userId).
		Delete(&PostCollection{}).Error
}

func (dao *PostCollectionDAO) Find(ctx context.Context, postId, userId uint64) (*PostCollection, error) {
	var collection PostCollection
	err := dao.db.WithContext(ctx).
		Where("post_id = ? AND user_id = ?", postId, userId).
		First(&collection).Error
	if err != nil {
		return nil, err
	}
	return &collection, nil
}

func (dao *PostCollectionDAO) Exists(ctx context.Context, postId, userId uint64) (bool, error) {
	var count int64
	err := dao.db.WithContext(ctx).
		Model(&PostCollection{}).
		Where("post_id = ? AND user_id = ?", postId, userId).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (dao *PostCollectionDAO) BatchCheckStatus(ctx context.Context, userId uint64, postIds []uint64) (map[uint64]bool, error) {
	var collections []PostCollection
	err := dao.db.WithContext(ctx).
		Where("user_id = ? AND post_id IN ?", userId, postIds).
		Find(&collections).Error
	if err != nil {
		return nil, err
	}

	statusMap := make(map[uint64]bool, len(postIds))
	for _, id := range postIds {
		statusMap[id] = false
	}
	for _, c := range collections {
		statusMap[c.PostID] = true
	}
	return statusMap, nil
}

func (dao *PostCollectionDAO) BatchCreate(ctx context.Context, collections []*PostCollection) error {
	if len(collections) == 0 {
		return nil
	}
	return dao.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(collections, 100).Error
}

func (dao *PostCollectionDAO) GetDB() *gorm.DB {
	return dao.db
}
