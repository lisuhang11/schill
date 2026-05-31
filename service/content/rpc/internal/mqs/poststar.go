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

type PostStarConsumer struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPostStarConsumer(ctx context.Context, svcCtx *svc.ServiceContext) *PostStarConsumer {
	return &PostStarConsumer{ctx: ctx, svcCtx: svcCtx}
}

func (c *PostStarConsumer) StartConsume(brokers []string, topic string, group string) error {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest
	consumerGroup, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		return err
	}
	defer consumerGroup.Close()

	handler := &postStarConsumerGroupHandler{ctx: c.ctx, svcCtx: c.svcCtx}
	for {
		if err := consumerGroup.Consume(c.ctx, []string{topic}, handler); err != nil {
			return err
		}
		if c.ctx.Err() != nil {
			return c.ctx.Err()
		}
	}
}

type postStarConsumerGroupHandler struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func (h *postStarConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *postStarConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *postStarConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var payload mq.PostStarMessage
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
				Where("id = ?", payload.PostID).
				Update("upvote_count", gorm.Expr("upvote_count + ?", 1)).Error
		})
		if err != nil {
			logx.Errorf("increment post star count failed: %v", err)
		}

		// Invalidate post caches so the next detail request rebuilds from DB with latest stats.
		_ = h.svcCtx.Redis.Del(h.ctx, fmt.Sprintf("%s%d", redis.PostStatsKey, payload.PostID))
		_ = h.svcCtx.Redis.Del(h.ctx, fmt.Sprintf("%s%d", redis.PostBaseKey, payload.PostID))
		_, _ = h.svcCtx.Redis.Incr(h.ctx, fmt.Sprintf("%s%d", redis.PostCacheVersionKey, payload.PostID))
		// Also bump the post author's list cache version.
		invalidatePostListForPost(h.svcCtx, payload.PostID)
		session.MarkMessage(msg, "")
	}
	return nil
}
