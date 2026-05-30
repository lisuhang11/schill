package mqs

import (
	"context"

	"SChill/common/mq"
	"SChill/service/content/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

func skipContentEvent(svcCtx *svc.ServiceContext, group string, envelope *mq.EventEnvelope) bool {
	key := mq.BuildIdempotencyKey(group, envelope)
	if key == "" {
		return false
	}
	ok, err := svcCtx.Redis.SetNX(context.Background(), key, "1", mq.DefaultEventTTL)
	if err != nil {
		logx.Errorf("content consumer idempotency check failed: %v", err)
		return false
	}
	return !ok
}
