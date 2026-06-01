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

type PostStarHandler struct {
	svcCtx *svc.ServiceContext
}

func NewPostStarHandler(svcCtx *svc.ServiceContext) *PostStarHandler {
	return &PostStarHandler{svcCtx: svcCtx}
}

func (h *PostStarHandler) Handle(ctx context.Context, msg *sarama.ConsumerMessage, envelope *mq.EventEnvelope) error {
	var payload mq.PostStarMessage
	if envelope != nil && len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, &payload); err != nil {
			logx.Errorf("decode post star data failed: %v", err)
			return err
		}
	} else {
		return fmt.Errorf("no envelope data in post star message")
	}

	err := h.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return tx.WithContext(ctx).Model(&model.Post{}).
			Where("id = ?", payload.PostID).
			Update("upvote_count", gorm.Expr("upvote_count + ?", 1)).Error
	})
	if err != nil {
		logx.Errorf("increment post star count failed: %v", err)
		return err
	}

	_ = h.svcCtx.Redis.Del(ctx, fmt.Sprintf("%s%d", redis.PostStatsKey, payload.PostID))
	_ = h.svcCtx.Redis.Del(ctx, fmt.Sprintf("%s%d", redis.PostBaseKey, payload.PostID))
	_, _ = h.svcCtx.Redis.Incr(ctx, fmt.Sprintf("%s%d", redis.PostCacheVersionKey, payload.PostID))
	invalidatePostListForPost(h.svcCtx, payload.PostID)
	return nil
}
