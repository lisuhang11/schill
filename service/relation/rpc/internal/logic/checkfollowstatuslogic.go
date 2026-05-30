package logic

import (
	"context"

	errutil "SChill/common/error"
	"SChill/service/relation/rpc/internal/svc"
	"SChill/service/relation/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CheckFollowStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCheckFollowStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckFollowStatusLogic {
	return &CheckFollowStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CheckFollowStatusLogic) CheckFollowStatus(in *pb.CheckFollowStatusReq) (*pb.CheckFollowStatusResp, error) {
	isFollow, err := relationExists(l.ctx, l.svcCtx, in.UserId, in.TargetUserId)
	if err != nil {
		logx.Errorf("check follow status failed: userId=%d targetUserId=%d err=%v", in.UserId, in.TargetUserId, err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	isMutual := false
	if isFollow {
		isMutual, err = relationExists(l.ctx, l.svcCtx, in.TargetUserId, in.UserId)
		if err != nil {
			logx.Errorf("check mutual follow status failed: userId=%d targetUserId=%d err=%v", in.UserId, in.TargetUserId, err)
			return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
		}
	}

	return &pb.CheckFollowStatusResp{
		IsFollow: isFollow,
		IsMutual: isMutual,
	}, nil
}
