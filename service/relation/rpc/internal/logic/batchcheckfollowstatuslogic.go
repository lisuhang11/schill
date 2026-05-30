package logic

import (
	"context"

	errutil "SChill/common/error"
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
	statusList := make([]*pb.FollowStatusItem, 0, len(in.TargetUserIds))
	for _, uid := range in.TargetUserIds {
		isFollow, err := relationExists(l.ctx, l.svcCtx, in.UserId, uid)
		if err != nil {
			logx.Errorf("batch check follow status failed: userId=%d targetUserId=%d err=%v", in.UserId, uid, err)
			return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
		}

		isMutual := false
		if isFollow {
			isMutual, err = relationExists(l.ctx, l.svcCtx, uid, in.UserId)
			if err != nil {
				logx.Errorf("batch check mutual status failed: userId=%d targetUserId=%d err=%v", in.UserId, uid, err)
				return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
			}
		}

		statusList = append(statusList, &pb.FollowStatusItem{
			UserId:   uid,
			IsFollow: isFollow,
			IsMutual: isMutual,
		})
	}

	return &pb.BatchCheckFollowStatusResp{
		Status: statusList,
	}, nil
}
