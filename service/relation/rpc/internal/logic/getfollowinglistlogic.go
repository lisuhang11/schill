package logic

import (
	"context"

	errutil "SChill/common/error"
	"SChill/service/relation/rpc/internal/model"
	"SChill/service/relation/rpc/internal/svc"
	"SChill/service/relation/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFollowingListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetFollowingListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFollowingListLogic {
	return &GetFollowingListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetFollowingListLogic) GetFollowingList(in *pb.GetFollowingListReq) (*pb.GetFollowingListResp, error) {
	var total int64
	err := l.svcCtx.DB.WithContext(l.ctx).
		Model(&model.Following{}).
		Where("user_id = ?", in.UserId).
		Count(&total).Error
	if err != nil {
		logx.Errorf("统计关注数量失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	if total == 0 {
		return &pb.GetFollowingListResp{
			Total: 0,
			List:  []*pb.FollowInfo{},
		}, nil
	}

	offset := (in.Page - 1) * in.PageSize
	var followings []model.Following
	err = l.svcCtx.DB.WithContext(l.ctx).
		Where("user_id = ?", in.UserId).
		Limit(int(in.PageSize)).
		Offset(int(offset)).
		Order("created_at DESC").
		Find(&followings).Error
	if err != nil {
		logx.Errorf("查询关注列表失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	userIds := make([]uint64, 0, len(followings))
	for _, f := range followings {
		userIds = append(userIds, f.FollowID)
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
		if user, ok := userMap[f.FollowID]; ok {
			list = append(list, &pb.FollowInfo{
				UserId:     user.ID,
				Username:   user.Username,
				Avatar:     user.Avatar,
				FollowTime: f.CreatedAt.Unix(),
			})
		}
	}

	return &pb.GetFollowingListResp{
		Total: total,
		List:  list,
	}, nil
}
