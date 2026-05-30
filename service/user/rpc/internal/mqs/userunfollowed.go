package mqs

import (
	"context"
	"encoding/json"

	"SChill/common/mq"
	"SChill/service/user/rpc/internal/logic"
	"SChill/service/user/rpc/internal/model"
	"SChill/service/user/rpc/internal/svc"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UserUnfollowedConsumer struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserUnfollowedConsumer(ctx context.Context, svcCtx *svc.ServiceContext) *UserUnfollowedConsumer {
	return &UserUnfollowedConsumer{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (c *UserUnfollowedConsumer) StartConsume(brokers []string, topic string, group string) error {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumerGroup, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		return err
	}
	defer consumerGroup.Close()

	handler := &userUnfollowedConsumerGroupHandler{
		ctx:    c.ctx,
		svcCtx: c.svcCtx,
	}

	for {
		if err := consumerGroup.Consume(c.ctx, []string{topic}, handler); err != nil {
			logx.Errorf("消费取消关注Kafka消息失败: %v", err)
			return err
		}
		if c.ctx.Err() != nil {
			return c.ctx.Err()
		}
	}
}

type userUnfollowedConsumerGroupHandler struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func (h *userUnfollowedConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *userUnfollowedConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *userUnfollowedConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		logx.Infof("收到用户取消关注消息: Topic=%s, Partition=%d, Offset=%d, Key=%s, Value=%s",
			msg.Topic, msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))

		var m mq.UserUnfollowedMessage
		if err := json.Unmarshal(msg.Value, &m); err != nil {
			logx.Errorf("解析取消关注消息失败: %v", err)
			session.MarkMessage(msg, "")
			continue
		}

		err := h.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
			// 更新取消关注者的关注数（减1，确保不小于0）
			var followerStat model.UserStat
			err := tx.WithContext(h.ctx).Where("user_id = ?", m.FollowerID).First(&followerStat).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					// 记录不存在，不需要处理
				} else {
					return err
				}
			} else {
				if err := tx.WithContext(h.ctx).Model(&followerStat).Update("following_count", gorm.Expr("GREATEST(following_count - ?, 0)", 1)).Error; err != nil {
					return err
				}
			}

			// 更新被取消关注者的粉丝数（减1，确保不小于0）
			var followingStat model.UserStat
			err = tx.WithContext(h.ctx).Where("user_id = ?", m.FollowingID).First(&followingStat).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					// 记录不存在，不需要处理
				} else {
					return err
				}
			} else {
				if err := tx.WithContext(h.ctx).Model(&followingStat).Update("follower_count", gorm.Expr("GREATEST(follower_count - ?, 0)", 1)).Error; err != nil {
					return err
				}
			}

			return nil
		})

		if err != nil {
			logx.Errorf("更新关注数和粉丝数失败: %v", err)
		}

		logic.InvalidateUserCaches(h.ctx, h.svcCtx, m.FollowerID)
		logic.InvalidateUserCaches(h.ctx, h.svcCtx, m.FollowingID)
		session.MarkMessage(msg, "")
	}
	return nil
}
