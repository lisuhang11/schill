package mqs

import (
	"context"
	"fmt"

	"SChill/common/mq"
	"SChill/common/redis"
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
	return &CommentDeletedConsumer{ctx: ctx, svcCtx: svcCtx}
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

	handler := &commentDeletedConsumerGroupHandler{ctx: c.ctx, svcCtx: c.svcCtx}
	for {
		if err := consumerGroup.Consume(c.ctx, []string{topic}, handler); err != nil {
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
		var payload mq.CommentDeletedMessage
		envelope, err := mq.DecodeEnvelopePayload(msg.Value, &payload)
		if err != nil {
			session.MarkMessage(msg, "")
			continue
		}
		if skipContentEvent(h.svcCtx, h.svcCtx.Config.KqConsumerConf.Group, envelope) {
			session.MarkMessage(msg, "")
			continue
		}

		err = h.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
			return tx.WithContext(h.ctx).Model(&model.Post{}).
				Where("id = ? AND comment_count > 0", payload.PostID).
				Update("comment_count", gorm.Expr("comment_count - ?", 1)).Error
		})
		if err != nil {
			logx.Errorf("decrement post comment count failed: %v", err)
		}

		_, _ = h.svcCtx.Redis.Incr(h.ctx, fmt.Sprintf("%s%d", redis.PostCacheVersionKey, payload.PostID))
		session.MarkMessage(msg, "")
	}
	return nil
}
