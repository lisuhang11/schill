package logic

import (
	"context"
	"strconv"

	errutil "SChill/common/error"
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
	total, err := relationCount(l.ctx, l.svcCtx, relationCacheFollowers, in.UserId)
	if err != nil {
		logx.Errorf("count followers failed: userId=%d err=%v", in.UserId, err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	if total == 0 {
		return &pb.GetFollowerListResp{
			Total: 0,
			List:  []*pb.FollowInfo{},
		}, nil
	}

	pairs, err := relationMembersWithScores(l.ctx, l.svcCtx, relationCacheFollowers, in.UserId, in.Page, in.PageSize)
	if err != nil {
		logx.Errorf("get follower cache list failed: userId=%d err=%v", in.UserId, err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	userIDs := make([]uint64, 0, len(pairs))
	for _, pair := range pairs {
		id, parseErr := strconv.ParseUint(pair.Key, 10, 64)
		if parseErr != nil {
			logx.Errorf("parse follower user id failed: key=%s err=%v", pair.Key, parseErr)
			continue
		}
		userIDs = append(userIDs, id)
	}

	userMap, err := batchLoadRelationUsers(l.ctx, l.svcCtx.UserRpc, userIDs)
	if err != nil {
		logx.Errorf("load follower user profiles failed: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	list := make([]*pb.FollowInfo, 0, len(pairs))
	for _, pair := range pairs {
		id, parseErr := strconv.ParseUint(pair.Key, 10, 64)
		if parseErr != nil {
			continue
		}
		user, ok := userMap[id]
		if !ok {
			continue
		}
		list = append(list, &pb.FollowInfo{
			UserId:     user.UserID,
			Username:   user.Username,
			Avatar:     user.Avatar,
			FollowTime: pair.Score,
		})
	}

	return &pb.GetFollowerListResp{
		Total: total,
		List:  list,
	}, nil
}
