package logic

import (
	errutil "SChill/common/error"
	"SChill/service/user/api/internal/svc"
	"SChill/service/user/api/internal/types"
	"SChill/service/user/rpc/pb"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserStatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserStatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserStatLogic {
	return &GetUserStatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserStatLogic) GetUserStat(req *types.GetUserStatReq) (resp *types.GetUserStatResp, err error) {
	rpcResp, err := l.svcCtx.UserRpc.GetUserStat(l.ctx, &pb.GetUserStatReq{
		UserId: req.UserId,
	})
	if err != nil {
		logx.Errorf("用户服务:调用 RPC GetUserStat 失败: %v", err)
		code, msg := errutil.ParseRpcError(err)
		return &types.GetUserStatResp{
			Code: code,
			Msg:  msg,
			Stat: types.UserStatInfo{},
		}, nil
	}

	return &types.GetUserStatResp{
		Code: errutil.Success,
		Msg:  errutil.GetCodeMessage(errutil.Success),
		Stat: types.UserStatInfo{
			UserId:          rpcResp.Stat.UserId,
			PostCount:       rpcResp.Stat.PostCount,
			CommentCount:    rpcResp.Stat.CommentCount,
			FollowerCount:   rpcResp.Stat.FollowerCount,
			FollowingCount:  rpcResp.Stat.FollowingCount,
			LikeCount:       rpcResp.Stat.LikeCount,
			CollectionCount: rpcResp.Stat.CollectionCount,
			LastActiveTime:  rpcResp.Stat.LastActiveTime,
		},
	}, nil
}
