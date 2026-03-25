package logic

import (
	"context"

	errutil "SChill/common/error"
	"SChill/service/relation/rpc/internal/model"
	"SChill/service/relation/rpc/internal/svc"
	"SChill/service/relation/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFollowerListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFollowerListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFollowerListLogic {
	return &GetFollowerListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetFollowerListLogic) GetFollowerList(in *pb.GetFollowerListReq) (*pb.GetFollowerListResp, error) {
	var total int64
	err := l.svcCtx.DB.WithContext(l.ctx).
		Model(&model.Following{}).
		Where("follow_id = ?", in.UserId).
		Count(&total).Error
	if err != nil {
		logx.Errorf("统计粉丝数量失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	if total == 0 {
		return &pb.GetFollowerListResp{
			Total: 0,
			List:  []*pb.FollowInfo{},
		}, nil
	}

	offset := (in.Page - 1) * in.PageSize
	var followings []model.Following
	err = l.svcCtx.DB.WithContext(l.ctx).
		Where("follow_id = ?", in.UserId).
		Limit(int(in.PageSize)).
		Offset(int(offset)).
		Order("created_at DESC").
		Find(&followings).Error
	if err != nil {
		logx.Errorf("查询粉丝列表失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	userIds := make([]uint64, 0, len(followings))
	for _, f := range followings {
		userIds = append(userIds, f.UserID)
	}

	var users []model.User
	err = l.svcCtx.DB.WithContext(l.ctx).
		Where("id IN ?", userIds).
		Find(&users).Error
	if err != nil {
		logx.Errorf("查询用户信息失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	userMap := make(map[uint64]model.User)
	for _, u := range users {
		userMap[u.ID] = u
	}

	var list []*pb.FollowInfo
	for _, f := range followings {
		if user, ok := userMap[f.UserID]; ok {
			list = append(list, &pb.FollowInfo{
				UserId:     user.ID,
				Username:   user.Username,
				Avatar:     user.Avatar,
				FollowTime: f.CreatedAt.Unix(),
			})
		}
	}

	return &pb.GetFollowerListResp{
		Total: total,
		List:  list,
	}, nil
}
