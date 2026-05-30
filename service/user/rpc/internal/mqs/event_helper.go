package mqs

import (
	"context"

	"SChill/common/mq"
	"SChill/service/user/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

func skipUserEvent(svcCtx *svc.ServiceContext, group string, envelope *mq.EventEnvelope) bool {
	key := mq.BuildIdempotencyKey(group, envelope)
	if key == "" {
		return false
	}
	ok, err := svcCtx.Redis.SetnxExCtx(context.Background(), key, "1", int(mq.DefaultEventTTL.Seconds()))
	if err != nil {
		logx.Errorf("user consumer idempotency check failed: %v", err)
		return false
	}
	return !ok
}
