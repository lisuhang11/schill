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

type CheckPostStarLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckPostStarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckPostStarLogic {
	return &CheckPostStarLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CheckPostStarLogic) CheckPostStar(in *pb.CheckPostStarReq) (*pb.CheckPostStarResp, error) {
	relationKey := fmt.Sprintf("%s%d", redis.PostLikeRelationKey, in.PostId)
	countKey := fmt.Sprintf("%s%d", redis.PostLikeCountKey, in.PostId)
	userIdStr := strconv.FormatUint(in.UserId, 10)

	isStarred, err := l.svcCtx.RedisClient.SIsMember(l.ctx, relationKey, userIdStr)
	if err != nil {
		logx.Errorf("Redis SIsMember query failed: %v", err)
		// Fall through to DB fallback
	} else {
		// Check whether the Redis key still exists.
		// SISMEMBER returns (false, nil) for non-existent keys, so we can't
		// distinguish "key expired" from "member not in set" without Exists.
		keyExists, existsErr := l.svcCtx.RedisClient.Exists(l.ctx, relationKey)
		if existsErr != nil {
			logx.Errorf("Redis Exists check failed: %v", existsErr)
			// Fall through to DB fallback on uncertainty
		} else if keyExists > 0 {
			// Redis key is alive — trust the cached result
			countStr, countErr := l.svcCtx.RedisClient.Get(l.ctx, countKey)
			if countErr != nil {
				logx.Errorf("Redis GET query failed: %v", countErr)
			}

			var starCount int64
			if countStr != "" {
				starCount, _ = strconv.ParseInt(countStr, 10, 64)
			}

			return &pb.CheckPostStarResp{
				IsStarred: isStarred,
				StarCount: starCount,
			}, nil
		}
		// Key does not exist (TTL expired) — fall through to DB fallback
	}

	// Redis miss (key expired or unavailable) — fallback to DB and backfill Redis
	exists, dbErr := l.svcCtx.PostStarDAO.Exists(l.ctx, in.PostId, in.UserId)
	if dbErr != nil {
		logx.Errorf("DB fallback for star check failed: postId=%d userId=%d err=%v", in.PostId, in.UserId, dbErr)
		return &pb.CheckPostStarResp{
			IsStarred: false,
			StarCount: 0,
		}, nil
	}

	if exists {
		// Backfill Redis asynchronously
		go func() {
			bgCtx := context.Background()
			ttl := redis.InteractionDefaultTTL
			ttlStr := strconv.Itoa(int(ttl))
			if _, evalErr := l.svcCtx.RedisClient.Eval(bgCtx, l.svcCtx.LuaScripts.PostLike,
				[]string{relationKey, countKey},
				userIdStr,
				ttlStr,
			); evalErr != nil {
				logx.Errorf("backfill star to Redis failed: postId=%d userId=%d err=%v", in.PostId, in.UserId, evalErr)
			}
		}()

		// Count from DB
		var dbCount int64
		if err := l.svcCtx.DB.WithContext(l.ctx).Model(&struct{ ID uint64 }{}).
			Table("post_star").Where("post_id = ?", in.PostId).Count(&dbCount).Error; err != nil {
			logx.Errorf("DB count for star failed: postId=%d err=%v", in.PostId, err)
		}

		return &pb.CheckPostStarResp{
			IsStarred: true,
			StarCount: dbCount,
		}, nil
	}

	return &pb.CheckPostStarResp{
		IsStarred: false,
		StarCount: 0,
	}, nil
}
