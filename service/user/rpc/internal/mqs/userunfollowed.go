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

type UserUnfollowedHandler struct {
	svcCtx *svc.ServiceContext
}

func NewUserUnfollowedHandler(svcCtx *svc.ServiceContext) *UserUnfollowedHandler {
	return &UserUnfollowedHandler{svcCtx: svcCtx}
}

func (h *UserUnfollowedHandler) Handle(ctx context.Context, msg *sarama.ConsumerMessage, envelope *mq.EventEnvelope) error {
	var m mq.UserUnfollowedMessage
	if err := json.Unmarshal(envelope.Data, &m); err != nil {
		logx.Errorf("解析取消关注消息失败: %v", err)
		return nil // skip malformed messages
	}

	err := h.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		// 更新取消关注者的关注数（减1，确保不小于0）
		var followerStat model.UserStat
		err := tx.WithContext(ctx).Where("user_id = ?", m.FollowerID).First(&followerStat).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// 记录不存在，不需要处理
			} else {
				return err
			}
		} else {
			if err := tx.WithContext(ctx).Model(&followerStat).Update("following_count", gorm.Expr("GREATEST(following_count - ?, 0)", 1)).Error; err != nil {
				return err
			}
		}

		// 更新被取消关注者的粉丝数（减1，确保不小于0）
		var followingStat model.UserStat
		err = tx.WithContext(ctx).Where("user_id = ?", m.FollowingID).First(&followingStat).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// 记录不存在，不需要处理
			} else {
				return err
			}
		} else {
			if err := tx.WithContext(ctx).Model(&followingStat).Update("follower_count", gorm.Expr("GREATEST(follower_count - ?, 0)", 1)).Error; err != nil {
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

// Compile-time check: UserUnfollowedHandler implements kafka.ConsumerHandler.
var _ kafka.ConsumerHandler = (*UserUnfollowedHandler)(nil)
