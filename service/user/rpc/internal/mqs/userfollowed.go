package mqs

import (
	"context"
	"encoding/json"

	"SChill/common/kafka"
	"SChill/common/mq"
	"SChill/service/user/rpc/internal/logic"
	"SChill/service/user/rpc/internal/model"
	"SChill/service/user/rpc/internal/svc"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UserFollowedHandler struct {
	svcCtx *svc.ServiceContext
}

func NewUserFollowedHandler(svcCtx *svc.ServiceContext) *UserFollowedHandler {
	return &UserFollowedHandler{svcCtx: svcCtx}
}

func (h *UserFollowedHandler) Handle(ctx context.Context, msg *sarama.ConsumerMessage, envelope *mq.EventEnvelope) error {
	var m mq.UserFollowedMessage
	if err := json.Unmarshal(envelope.Data, &m); err != nil {
		logx.Errorf("解析关注消息失败: %v", err)
		return nil // skip malformed messages
	}

	err := h.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		// 更新关注者的关注数
		var followerStat model.UserStat
		err := tx.WithContext(ctx).Where("user_id = ?", m.FollowerID).First(&followerStat).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				followerStat = model.UserStat{
					UserID:         m.FollowerID,
					FollowingCount: 1,
				}
				if err := tx.WithContext(ctx).Create(&followerStat).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if err := tx.WithContext(ctx).Model(&followerStat).Update("following_count", gorm.Expr("following_count + ?", 1)).Error; err != nil {
				return err
			}
		}

		// 更新被关注者的粉丝数
		var followingStat model.UserStat
		err = tx.WithContext(ctx).Where("user_id = ?", m.FollowingID).First(&followingStat).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				followingStat = model.UserStat{
					UserID:        m.FollowingID,
					FollowerCount: 1,
				}
				if err := tx.WithContext(ctx).Create(&followingStat).Error; err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			if err := tx.WithContext(ctx).Model(&followingStat).Update("follower_count", gorm.Expr("follower_count + ?", 1)).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		logx.Errorf("更新关注数和粉丝数失败: %v", err)
		return err
	}

	logic.InvalidateUserCaches(ctx, h.svcCtx, m.FollowerID)
	logic.InvalidateUserCaches(ctx, h.svcCtx, m.FollowingID)
	return nil
}

// Compile-time check: UserFollowedHandler implements kafka.ConsumerHandler.
var _ kafka.ConsumerHandler = (*UserFollowedHandler)(nil)
