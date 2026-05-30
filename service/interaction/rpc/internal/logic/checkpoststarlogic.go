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
		logx.Errorf("Redis SIsMember 查询失败: %v", err)
	}

	countStr, err := l.svcCtx.RedisClient.Get(l.ctx, countKey)
	if err != nil {
		logx.Errorf("Redis GET 查询失败: %v", err)
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
