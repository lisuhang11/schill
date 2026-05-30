package model

import (
	"context"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PostStarDAO struct {
	db *gorm.DB
}

func NewPostStarDAO(db *gorm.DB) *PostStarDAO {
	return &PostStarDAO{db: db}
}

func (dao *PostStarDAO) Create(ctx context.Context, star *PostStar) error {
	return dao.db.WithContext(ctx).Create(star).Error
}

func (dao *PostStarDAO) Delete(ctx context.Context, postId, userId uint64) error {
	return dao.db.WithContext(ctx).
		Where("post_id = ? AND user_id = ?", postId, userId).
		Delete(&PostStar{}).Error
}

func (dao *PostStarDAO) Find(ctx context.Context, postId, userId uint64) (*PostStar, error) {
	var star PostStar
	err := dao.db.WithContext(ctx).
		Where("post_id = ? AND user_id = ?", postId, userId).
		First(&star).Error
	if err != nil {
		return nil, err
	}
	return &star, nil
}

func (dao *PostStarDAO) Exists(ctx context.Context, postId, userId uint64) (bool, error) {
	var count int64
	err := dao.db.WithContext(ctx).
		Model(&PostStar{}).
		Where("post_id = ? AND user_id = ?", postId, userId).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (dao *PostStarDAO) BatchCheckStatus(ctx context.Context, userId uint64, postIds []uint64) (map[uint64]bool, error) {
	var stars []PostStar
	err := dao.db.WithContext(ctx).
		Where("user_id = ? AND post_id IN ?", userId, postIds).
		Find(&stars).Error
	if err != nil {
		return nil, err
	}

	statusMap := make(map[uint64]bool, len(postIds))
	for _, id := range postIds {
		statusMap[id] = false
	}
	for _, s := range stars {
		statusMap[s.PostID] = true
	}
	return statusMap, nil
}

func (dao *PostStarDAO) BatchCreate(ctx context.Context, stars []*PostStar) error {
	if len(stars) == 0 {
		return nil
	}
	return dao.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(stars, 100).Error
}

func (dao *PostStarDAO) GetDB() *gorm.DB {
	return dao.db
}
