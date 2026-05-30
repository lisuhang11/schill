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

// 检查动态收藏状态
func (l *CheckPostCollectionLogic) CheckPostCollection(in *pb.CheckPostCollectionReq) (*pb.CheckPostCollectionResp, error) {
	relationKey := fmt.Sprintf("%s%d", redis.PostCollectionRelationKey, in.PostId)
	userIdStr := strconv.FormatUint(in.UserId, 10)

	isCollected, err := l.svcCtx.RedisClient.SIsMember(l.ctx, relationKey, userIdStr)
	if err != nil {
		logx.Errorf("Redis SIsMember 查询失败: %v", err)
	}

	return &pb.CheckPostCollectionResp{
		IsCollected: isCollected,
	}, nil
}
