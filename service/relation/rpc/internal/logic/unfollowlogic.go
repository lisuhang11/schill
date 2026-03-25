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

type UnfollowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnfollowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnfollowLogic {
	return &UnfollowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UnfollowLogic) Unfollow(in *pb.UnfollowReq) (*pb.UnfollowResp, error) {
	var following model.Following
	err := l.svcCtx.DB.WithContext(l.ctx).
		Where("user_id = ? AND follow_id = ?", in.UserId, in.TargetUserId).
		First(&following).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errutil.RpcBusinessError(errutil.ErrNotFollowing)
		}
		logx.Errorf("查询关注关系失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	if err := l.svcCtx.DB.WithContext(l.ctx).Delete(&following).Error; err != nil {
		logx.Errorf("删除关注关系失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	return &pb.UnfollowResp{
		Success: true,
	}, nil
}
