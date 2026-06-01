package mqs

import (
	"context"
	"encoding/json"
	"fmt"

	"SChill/common/mq"
	"SChill/common/redis"
	"SChill/service/content/rpc/internal/model"
	"SChill/service/content/rpc/internal/svc"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type PostUncollectHandler struct {
	svcCtx *svc.ServiceContext
}

func NewPostUncollectHandler(svcCtx *svc.ServiceContext) *PostUncollectHandler {
	return &PostUncollectHandler{svcCtx: svcCtx}
}

func (h *PostUncollectHandler) Handle(ctx context.Context, msg *sarama.ConsumerMessage, envelope *mq.EventEnvelope) error {
	var payload mq.PostUncollectMessage
	if envelope != nil && len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, &payload); err != nil {
			logx.Errorf("decode post uncollect data failed: %v", err)
			return err
		}
	} else {
		return fmt.Errorf("no envelope data in post uncollect message")
	}

	err := h.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return tx.WithContext(ctx).Model(&model.Post{}).
			Where("id = ? AND collection_count > 0", payload.PostID).
			Update("collection_count", gorm.Expr("collection_count - ?", 1)).Error
	})
	if err != nil {
		logx.Errorf("decrement post collection count failed: %v", err)
		return err
	}

	_ = h.svcCtx.Redis.Del(ctx, fmt.Sprintf("%s%d", redis.PostStatsKey, payload.PostID))
	_ = h.svcCtx.Redis.Del(ctx, fmt.Sprintf("%s%d", redis.PostBaseKey, payload.PostID))
	_, _ = h.svcCtx.Redis.Incr(ctx, fmt.Sprintf("%s%d", redis.PostCacheVersionKey, payload.PostID))
	invalidatePostListForPost(h.svcCtx, payload.PostID)
	return nil
}
