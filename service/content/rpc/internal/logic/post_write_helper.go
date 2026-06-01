package logic

import (
	"context"
	"time"

	errutil "SChill/common/error"
	"SChill/service/content/rpc/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func loadOwnedPost(ctx context.Context, db *gorm.DB, postID, userID uint64) (*model.Post, error) {
	var post model.Post
	if err := db.WithContext(ctx).Where("id = ? AND deleted_at IS NULL", postID).First(&post).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errutil.RpcBusinessError(errutil.ErrPostNotExist)
		}
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}
	if post.UserID != userID {
		return nil, errutil.RpcBusinessError(errutil.ErrNoPermission)
	}

	return &post, nil
}

func replacePostContentsTx(ctx context.Context, tx *gorm.DB, postID, userID uint64, contents []*model.PostContent) error {
	now := time.Now()
	if err := tx.WithContext(ctx).
		Model(&model.PostContent{}).
		Where("post_id = ? AND deleted_at IS NULL", postID).
		Update("deleted_at", &now).Error; err != nil {
		return err
	}

	if len(contents) == 0 {
		return nil
	}

	for i := range contents {
		contents[i].PostID = postID
		contents[i].UserID = userID
	}

	return tx.WithContext(ctx).Create(&contents).Error
}

func replacePostTopicsTx(ctx context.Context, tx *gorm.DB, postID uint64, topicNames []string) error {
	var existing []model.PostTopic
	if err := tx.WithContext(ctx).Where("post_id = ?", postID).Find(&existing).Error; err != nil {
		return err
	}

	if len(existing) > 0 {
		topicIDs := make([]uint64, 0, len(existing))
		for _, item := range existing {
			topicIDs = append(topicIDs, item.TopicID)
		}

		if err := tx.WithContext(ctx).
			Model(&model.Topic{}).
			Where("id IN ? AND quote_num > 0", topicIDs).
			Update("quote_num", gorm.Expr("quote_num - 1")).Error; err != nil {
			return err
		}

		if err := tx.WithContext(ctx).Where("post_id = ?", postID).Delete(&model.PostTopic{}).Error; err != nil {
			return err
		}
	}

	for _, topicName := range normalizeTopicNames(topicNames) {
		topic := model.Topic{Name: topicName}
		if err := tx.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&topic).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("name = ? AND deleted_at IS NULL", topicName).First(&topic).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&model.PostTopic{
			PostID:  postID,
			TopicID: topic.ID,
		}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).
			Model(&model.Topic{}).
			Where("id = ?", topic.ID).
			Update("quote_num", gorm.Expr("quote_num + 1")).Error; err != nil {
			return err
		}
	}

	return nil
}

