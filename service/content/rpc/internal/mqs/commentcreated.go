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

type CommentCreatedHandler struct {
	svcCtx *svc.ServiceContext
}

func NewCommentCreatedHandler(svcCtx *svc.ServiceContext) *CommentCreatedHandler {
	return &CommentCreatedHandler{svcCtx: svcCtx}
}

func (h *CommentCreatedHandler) Handle(ctx context.Context, msg *sarama.ConsumerMessage, envelope *mq.EventEnvelope) error {
	var payload mq.CommentCreatedMessage
	if envelope != nil && len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, &payload); err != nil {
			logx.Errorf("decode comment created data failed: %v", err)
			return err
		}
	} else {
		return fmt.Errorf("no envelope data in comment created message")
	}

	err := h.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		return tx.WithContext(ctx).Model(&model.Post{}).
			Where("id = ?", payload.PostID).
			Update("comment_count", gorm.Expr("comment_count + ?", 1)).Error
	})
	if err != nil {
		logx.Errorf("increment post comment count failed: %v", err)
		return err
	}

	_ = h.svcCtx.Redis.Del(ctx, fmt.Sprintf("%s%d", redis.PostStatsKey, payload.PostID))
	_ = h.svcCtx.Redis.Del(ctx, fmt.Sprintf("%s%d", redis.PostBaseKey, payload.PostID))
	_, _ = h.svcCtx.Redis.Incr(ctx, fmt.Sprintf("%s%d", redis.PostCacheVersionKey, payload.PostID))
	invalidatePostListForPost(h.svcCtx, payload.PostID)
	return nil
}
