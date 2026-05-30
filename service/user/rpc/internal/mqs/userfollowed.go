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

type UserFollowedConsumer struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUserFollowedConsumer(ctx context.Context, svcCtx *svc.ServiceContext) *UserFollowedConsumer {
	return &UserFollowedConsumer{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (c *UserFollowedConsumer) StartConsume(brokers []string, topic string, group string) error {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumerGroup, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		return err
	}
	defer consumerGroup.Close()

	handler := &userFollowedConsumerGroupHandler{
		ctx:    c.ctx,
		svcCtx: c.svcCtx,
	}

	for {
		if err := consumerGroup.Consume(c.ctx, []string{topic}, handler); err != nil {
			logx.Errorf("消费关注Kafka消息失败: %v", err)
			return err
		}
		if c.ctx.Err() != nil {
			return c.ctx.Err()
		}
	}
}

type userFollowedConsumerGroupHandler struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func (h *userFollowedConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *userFollowedConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *userFollowedConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		logx.Infof("收到用户关注消息: Topic=%s, Partition=%d, Offset=%d, Key=%s, Value=%s",
			msg.Topic, msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))

		var m mq.UserFollowedMessage
		if err := json.Unmarshal(msg.Value, &m); err != nil {
			logx.Errorf("解析关注消息失败: %v", err)
			session.MarkMessage(msg, "")
			continue
		}

		err := h.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
			// 更新关注者的关注数
			var followerStat model.UserStat
			err := tx.WithContext(h.ctx).Where("user_id = ?", m.FollowerID).First(&followerStat).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					followerStat = model.UserStat{
						UserID:         m.FollowerID,
						FollowingCount: 1,
					}
					if err := tx.WithContext(h.ctx).Create(&followerStat).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			} else {
				if err := tx.WithContext(h.ctx).Model(&followerStat).Update("following_count", gorm.Expr("following_count + ?", 1)).Error; err != nil {
					return err
				}
			}

			// 更新被关注者的粉丝数
			var followingStat model.UserStat
			err = tx.WithContext(h.ctx).Where("user_id = ?", m.FollowingID).First(&followingStat).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					followingStat = model.UserStat{
						UserID:        m.FollowingID,
						FollowerCount: 1,
					}
					if err := tx.WithContext(h.ctx).Create(&followingStat).Error; err != nil {
						return err
					}
				} else {
					return err
				}
			} else {
				if err := tx.WithContext(h.ctx).Model(&followingStat).Update("follower_count", gorm.Expr("follower_count + ?", 1)).Error; err != nil {
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
