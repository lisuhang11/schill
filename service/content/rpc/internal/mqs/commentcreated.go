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

type CommentCreatedConsumer struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCommentCreatedConsumer(ctx context.Context, svcCtx *svc.ServiceContext) *CommentCreatedConsumer {
	return &CommentCreatedConsumer{ctx: ctx, svcCtx: svcCtx}
}

func (c *CommentCreatedConsumer) StartConsume(brokers []string, topic string, group string) error {
	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumerGroup, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		return err
	}
	defer consumerGroup.Close()

	handler := &commentCreatedConsumerGroupHandler{ctx: c.ctx, svcCtx: c.svcCtx}
	for {
		if err := consumerGroup.Consume(c.ctx, []string{topic}, handler); err != nil {
			logx.Errorf("consume comment created failed: %v", err)
			return err
		}
		if c.ctx.Err() != nil {
			return c.ctx.Err()
		}
	}
}

type commentCreatedConsumerGroupHandler struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func (h *commentCreatedConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *commentCreatedConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *commentCreatedConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var payload mq.CommentCreatedMessage
		envelope, err := mq.DecodeEnvelopePayload(msg.Value, &payload)
		if err != nil {
			logx.Errorf("decode comment created event failed: %v", err)
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
				Update("comment_count", gorm.Expr("comment_count + ?", 1)).Error
		})
		if err != nil {
			logx.Errorf("increment post comment count failed: %v", err)
		}

		// Comment count change invalidates post detail and list caches.
		_ = h.svcCtx.Redis.Del(h.ctx, fmt.Sprintf("%s%d", redis.PostStatsKey, payload.PostID))
		_ = h.svcCtx.Redis.Del(h.ctx, fmt.Sprintf("%s%d", redis.PostBaseKey, payload.PostID))
		_, _ = h.svcCtx.Redis.Incr(h.ctx, fmt.Sprintf("%s%d", redis.PostCacheVersionKey, payload.PostID))
		invalidatePostListForPost(h.svcCtx, payload.PostID)
		session.MarkMessage(msg, "")
	}
	return nil
}
