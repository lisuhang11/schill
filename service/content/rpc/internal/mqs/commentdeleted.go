package mqs

import (
	"context"
	"encoding/json"

	"SChill/common/mq"
	"SChill/service/content/rpc/internal/model"
	"SChill/service/content/rpc/internal/svc"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type CommentDeletedConsumer struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCommentDeletedConsumer(ctx context.Context, svcCtx *svc.ServiceContext) *CommentDeletedConsumer {
	return &CommentDeletedConsumer{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (c *CommentDeletedConsumer) StartConsume(brokers []string, topic string, group string) error {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumerGroup, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		return err
	}
	defer consumerGroup.Close()

	handler := &commentDeletedConsumerGroupHandler{
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

type commentDeletedConsumerGroupHandler struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func (h *commentDeletedConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *commentDeletedConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *commentDeletedConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		logx.Infof("收到评论删除消息: Topic=%s, Partition=%d, Offset=%d, Key=%s, Value=%s",
			msg.Topic, msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))

		var m mq.CommentDeletedMessage
		if err := json.Unmarshal(msg.Value, &m); err != nil {
			logx.Errorf("解析消息失败: %v", err)
			session.MarkMessage(msg, "")
			continue
		}

		err := h.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
			return tx.WithContext(h.ctx).Model(&model.Post{}).
				Where("id = ?", m.PostID).
				Update("comment_count", gorm.Expr("comment_count - ?", 1)).Error
		})

		if err != nil {
			logx.Errorf("减少帖子评论数失败: %v", err)
		}

		session.MarkMessage(msg, "")
	}
	return nil
}
