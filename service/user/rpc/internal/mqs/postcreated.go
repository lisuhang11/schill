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

type PostCreatedHandler struct {
	svcCtx *svc.ServiceContext
}

func NewPostCreatedHandler(svcCtx *svc.ServiceContext) *PostCreatedHandler {
	return &PostCreatedHandler{svcCtx: svcCtx}
}

func (h *PostCreatedHandler) Handle(ctx context.Context, msg *sarama.ConsumerMessage, envelope *mq.EventEnvelope) error {
	var m mq.PostCreatedMessage
	if err := json.Unmarshal(envelope.Data, &m); err != nil {
		logx.Errorf("解析帖子创建消息失败: %v", err)
		return nil // skip malformed messages
	}

	err := h.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		var stat model.UserStat
		err := tx.WithContext(ctx).Where("user_id = ?", m.UserID).First(&stat).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				stat = model.UserStat{
					UserID:    m.UserID,
					PostCount: 1,
				}
				return tx.WithContext(ctx).Create(&stat).Error
			}
			return err
		}
		return tx.WithContext(ctx).Model(&stat).Update("post_count", gorm.Expr("post_count + ?", 1)).Error
	})

	if err != nil {
		logx.Errorf("增加用户发帖数失败: %v", err)
		return err
	}

	logic.InvalidateUserCaches(ctx, h.svcCtx, m.UserID)
	return nil
}

// Compile-time check: PostCreatedHandler implements kafka.ConsumerHandler.
var _ kafka.ConsumerHandler = (*PostCreatedHandler)(nil)
