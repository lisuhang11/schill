package logic

import (
	"context"

	errutil "SChill/common/error"
	"SChill/service/relation/rpc/internal/model"
	"SChill/service/relation/rpc/internal/svc"
	"SChill/service/relation/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
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
	var following model.Following
	err := l.svcCtx.DB.WithContext(l.ctx).
		Where("user_id = ? AND follow_id = ?", in.UserId, in.TargetUserId).
		First(&following).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return &pb.CheckFollowStatusResp{
				IsFollow: false,
			}, nil
		}
		logx.Errorf("查询关注关系失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	return &pb.CheckFollowStatusResp{
		IsFollow: true,
	}, nil
}
