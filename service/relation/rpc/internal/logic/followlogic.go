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

type FollowLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewFollowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FollowLogic {
	return &FollowLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *FollowLogic) Follow(in *pb.FollowReq) (*pb.FollowResp, error) {
	if in.UserId == in.TargetUserId {
		return nil, errutil.RpcBusinessError(errutil.ErrCannotFollowSelf)
	}

	var existing model.Following
	err := l.svcCtx.DB.WithContext(l.ctx).
		Where("user_id = ? AND follow_id = ?", in.UserId, in.TargetUserId).
		First(&existing).Error
	if err == nil {
		return nil, errutil.RpcBusinessError(errutil.ErrAlreadyFollowed)
	}
	if err != gorm.ErrRecordNotFound {
		logx.Errorf("查询关注关系失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	following := model.Following{
		UserID:   in.UserId,
		FollowID: in.TargetUserId,
	}
	if err := l.svcCtx.DB.WithContext(l.ctx).Create(&following).Error; err != nil {
		logx.Errorf("创建关注关系失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	return &pb.FollowResp{
		Success: true,
	}, nil
}
