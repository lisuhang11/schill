package logic

import (
	"SChill/common/errorcode"
	"SChill/service/user/api/internal/svc"
	"SChill/service/user/api/internal/types"
	"SChill/service/user/rpc/pb"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	// 调用 RPC 获取用户统计
	rpcResp, err := l.svcCtx.UserRpc.GetUserStat(l.ctx, &pb.GetUserStatReq{
		UserId: req.UserId,
	})
	if err != nil {
		logx.Errorf("调用 RPC GetUserStat 失败: %v", err)
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.NotFound:
			return &types.GetUserStatResp{
				Code: errorcode.ErrUserNotExist,
				Msg:  errorcode.GetCodeMessage(errorcode.ErrUserNotExist),
			}, nil
		default:
			return &types.GetUserStatResp{
				Code: errorcode.ErrInternalError,
				Msg:  errorcode.GetCodeMessage(errorcode.ErrInternalError),
			}, nil
		}
	}

	stat := types.UserStatInfo{
		UserId:          rpcResp.Stat.UserId,
		PostCount:       rpcResp.Stat.PostCount,
		CommentCount:    rpcResp.Stat.CommentCount,
		FollowerCount:   rpcResp.Stat.FollowerCount,
		FollowingCount:  rpcResp.Stat.FollowingCount,
		LikeCount:       rpcResp.Stat.LikeCount,
		CollectionCount: rpcResp.Stat.CollectionCount,
		LastActiveTime:  rpcResp.Stat.LastActiveTime,
	}

	return &types.GetUserStatResp{
		Code: errorcode.Success,
		Msg:  errorcode.GetCodeMessage(errorcode.Success),
		Stat: stat,
	}, nil
}
