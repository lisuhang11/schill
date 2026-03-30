package mqs

import (
	"context"
	"encoding/json"

	"SChill/service/user/rpc/internal/model"
	"SChill/service/user/rpc/internal/svc"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type PostCreatedMessage struct {
	UserID uint64 `json:"user_id"`
	PostID uint64 `json:"post_id"`
}

type PostCreatedConsumer struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPostCreatedConsumer(ctx context.Context, svcCtx *svc.ServiceContext) *PostCreatedConsumer {
	return &PostCreatedConsumer{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (c *PostCreatedConsumer) StartConsume(brokers []string, topic string, group string) error {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumerGroup, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		return err
	}
	defer consumerGroup.Close()

	handler := &postCreatedConsumerGroupHandler{
		ctx:    c.ctx,
		svcCtx: c.svcCtx,
	}

	for {
		if err := consumerGroup.Consume(c.ctx, []string{topic}, handler); err != nil {
			logx.Errorf("消费Kafka消息失败: %v", err)
			return err
		}
		if c.ctx.Err() != nil {
			return c.ctx.Err()
		}
	}
}

type postCreatedConsumerGroupHandler struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func (h *postCreatedConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *postCreatedConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *postCreatedConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		logx.Infof("收到帖子创建消息: Topic=%s, Partition=%d, Offset=%d, Key=%s, Value=%s",
			msg.Topic, msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))

		var m PostCreatedMessage
		if err := json.Unmarshal(msg.Value, &m); err != nil {
			logx.Errorf("解析消息失败: %v", err)
			session.MarkMessage(msg, "")
			continue
		}

		err := h.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
			var stat model.UserStat
			err := tx.WithContext(h.ctx).Where("user_id = ?", m.UserID).First(&stat).Error
			if err != nil {
				if err == gorm.ErrRecordNotFound {
					stat = model.UserStat{
						UserID:    m.UserID,
						PostCount: 1,
					}
					return tx.WithContext(h.ctx).Create(&stat).Error
				}
				return err
			}
			return tx.WithContext(h.ctx).Model(&stat).Update("post_count", gorm.Expr("post_count + ?", 1)).Error
		})

		if err != nil {
			logx.Errorf("增加用户发帖数失败: %v", err)
		}

		session.MarkMessage(msg, "")
	}
	return nil
}
