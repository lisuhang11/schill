package logic

import (
	"context"

	errutil "SChill/common/error"
	"SChill/service/relation/rpc/internal/model"
	"SChill/service/relation/rpc/internal/svc"
	"SChill/service/relation/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchCheckFollowStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewBatchCheckFollowStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchCheckFollowStatusLogic {
	return &BatchCheckFollowStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *BatchCheckFollowStatusLogic) BatchCheckFollowStatus(in *pb.BatchCheckFollowStatusReq) (*pb.BatchCheckFollowStatusResp, error) {
	var followings []model.Following
	err := l.svcCtx.DB.WithContext(l.ctx).
		Where("user_id = ? AND follow_id IN ?", in.UserId, in.TargetUserIds).
		Find(&followings).Error
	if err != nil {
		logx.Errorf("批量查询关注关系失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	followMap := make(map[uint64]bool)
	for _, f := range followings {
		followMap[f.FollowID] = true
	}

	var statusList []*pb.FollowStatusItem
	for _, uid := range in.TargetUserIds {
		statusList = append(statusList, &pb.FollowStatusItem{
			UserId:   uid,
			IsFollow: followMap[uid],
		})
	}

	return &pb.BatchCheckFollowStatusResp{
		Status: statusList,
	}, nil
}
