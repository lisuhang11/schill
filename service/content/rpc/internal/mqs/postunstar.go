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

type PostUnstarConsumer struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPostUnstarConsumer(ctx context.Context, svcCtx *svc.ServiceContext) *PostUnstarConsumer {
	return &PostUnstarConsumer{ctx: ctx, svcCtx: svcCtx}
}

func (c *PostUnstarConsumer) StartConsume(brokers []string, topic string, group string) error {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	consumerGroup, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		return err
	}
	defer consumerGroup.Close()

	handler := &postUnstarConsumerGroupHandler{ctx: c.ctx, svcCtx: c.svcCtx}
	for {
		if err := consumerGroup.Consume(c.ctx, []string{topic}, handler); err != nil {
			return err
		}
		if c.ctx.Err() != nil {
			return c.ctx.Err()
		}
	}
}

type postUnstarConsumerGroupHandler struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func (h *postUnstarConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *postUnstarConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *postUnstarConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var payload mq.PostUnstarMessage
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
				Where("id = ? AND upvote_count > 0", payload.PostID).
				Update("upvote_count", gorm.Expr("upvote_count - ?", 1)).Error
		})
		if err != nil {
			logx.Errorf("decrement post star count failed: %v", err)
		}

		_ = h.svcCtx.Redis.Del(h.ctx, fmt.Sprintf("%s%d", redis.PostStatsKey, payload.PostID))
		_ = h.svcCtx.Redis.Del(h.ctx, fmt.Sprintf("%s%d", redis.PostBaseKey, payload.PostID))
		_, _ = h.svcCtx.Redis.Incr(h.ctx, fmt.Sprintf("%s%d", redis.PostCacheVersionKey, payload.PostID))
		invalidatePostListForPost(h.svcCtx, payload.PostID)
		session.MarkMessage(msg, "")
	}
	return nil
}
