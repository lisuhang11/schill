package mqs

import (
	"context"
	"encoding/json"

	"SChill/common/kafka"
	"SChill/common/mq"
	"SChill/service/user/rpc/internal/logic"
	"SChill/service/user/rpc/internal/model"
	"SChill/service/user/rpc/internal/svc"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type PostDeletedHandler struct {
	svcCtx *svc.ServiceContext
}

func NewPostDeletedHandler(svcCtx *svc.ServiceContext) *PostDeletedHandler {
	return &PostDeletedHandler{svcCtx: svcCtx}
}

func (h *PostDeletedHandler) Handle(ctx context.Context, msg *sarama.ConsumerMessage, envelope *mq.EventEnvelope) error {
	var m mq.PostDeletedMessage
	if err := json.Unmarshal(envelope.Data, &m); err != nil {
		logx.Errorf("解析帖子删除消息失败: %v", err)
		return nil // skip malformed messages
	}

	err := h.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		var stat model.UserStat
		err := tx.WithContext(ctx).Where("user_id = ?", m.UserID).First(&stat).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				stat = model.UserStat{
					UserID:    m.UserID,
					PostCount: 0,
				}
				return tx.WithContext(ctx).Create(&stat).Error
			}
			return err
		}

		if stat.PostCount > 0 {
			return tx.WithContext(ctx).Model(&stat).Update("post_count", gorm.Expr("post_count - ?", 1)).Error
		}
		return nil
	})

	if err != nil {
		logx.Errorf("减少用户发帖数失败: %v", err)
		return err
	}

	logic.InvalidateUserCaches(ctx, h.svcCtx, m.UserID)
	return nil
}

// Compile-time check: PostDeletedHandler implements kafka.ConsumerHandler.
var _ kafka.ConsumerHandler = (*PostDeletedHandler)(nil)
