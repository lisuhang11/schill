package logic

import (
	"context"
	"fmt"
	"strconv"

	"SChill/common/redis"
	"SChill/service/interaction/rpc/internal/svc"
	"SChill/service/interaction/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckPostCollectionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckPostCollectionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckPostCollectionLogic {
	return &CheckPostCollectionLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// CheckPostCollection checks if a user has collected a post.
// Falls back to DB when Redis key has expired (TTL-based eviction).
func (l *CheckPostCollectionLogic) CheckPostCollection(in *pb.CheckPostCollectionReq) (*pb.CheckPostCollectionResp, error) {
	relationKey := fmt.Sprintf("%s%d", redis.PostCollectionRelationKey, in.PostId)
	userIdStr := strconv.FormatUint(in.UserId, 10)

	isCollected, err := l.svcCtx.RedisClient.SIsMember(l.ctx, relationKey, userIdStr)
	if err != nil {
		logx.Errorf("Redis SIsMember query failed: %v", err)
	} else {
		// Check whether the Redis key still exists.
		// SISMEMBER returns (false, nil) for non-existent keys, so we can't
		// distinguish "key expired" from "member not in set" without Exists.
		keyExists, existsErr := l.svcCtx.RedisClient.Exists(l.ctx, relationKey)
		if existsErr != nil {
			logx.Errorf("Redis Exists check failed: %v", existsErr)
			// Fall through to DB fallback on uncertainty
		} else if keyExists > 0 {
			return &pb.CheckPostCollectionResp{
				IsCollected: isCollected,
			}, nil
		}
		// Key does not exist (TTL expired) — fall through to DB fallback
	}

	// Redis miss (key expired or unavailable) — fallback to DB and backfill Redis
	exists, dbErr := l.svcCtx.PostCollectionDAO.Exists(l.ctx, in.PostId, in.UserId)
	if dbErr != nil {
		logx.Errorf("DB fallback for collection check failed: postId=%d userId=%d err=%v", in.PostId, in.UserId, dbErr)
		return &pb.CheckPostCollectionResp{
			IsCollected: false,
		}, nil
	}

	if exists {
		go func() {
			bgCtx := context.Background()
			ttl := strconv.Itoa(int(redis.InteractionDefaultTTL))
			countKey := fmt.Sprintf("%s%d", redis.PostCollectionCountKey, in.PostId)
			if _, evalErr := l.svcCtx.RedisClient.Eval(bgCtx, l.svcCtx.LuaScripts.PostCollect,
				[]string{relationKey, countKey},
				userIdStr,
				ttl,
			); evalErr != nil {
				logx.Errorf("backfill collection to Redis failed: postId=%d userId=%d err=%v", in.PostId, in.UserId, evalErr)
			}
		}()

		return &pb.CheckPostCollectionResp{
			IsCollected: true,
		}, nil
	}

	return &pb.CheckPostCollectionResp{
		IsCollected: false,
	}, nil
}
