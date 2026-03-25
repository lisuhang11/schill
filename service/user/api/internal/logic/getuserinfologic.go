// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	errutil "SChill/common/error"
	"SChill/service/user/api/internal/svc"
	"SChill/service/user/api/internal/types"
	"SChill/service/user/rpc/pb"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserInfoLogic) GetUserInfo(req *types.GetUserInfoReq) (resp *types.GetUserInfoResp, err error) {
	rpcResp, err := l.svcCtx.UserRpc.GetUserInfo(l.ctx, &pb.GetUserInfoReq{
		UserId: req.UserId,
	})
	if err != nil {
		logx.Errorf("获取用户信息失败: %v", err)
		code, msg := errutil.ParseRpcError(err)
		return &types.GetUserInfoResp{
			Code:     code,
			Msg:      msg,
			UserInfo: types.UserInfo{},
		}, nil
	}

	return &types.GetUserInfoResp{
		Code: errutil.Success,
		Msg:  errutil.GetCodeMessage(errutil.Success),
		UserInfo: types.UserInfo{
			Id:        rpcResp.UserInfo.Id,
			Username:  rpcResp.UserInfo.Username,
			Nickname:  rpcResp.UserInfo.Nickname,
			Phone:     rpcResp.UserInfo.Phone,
			Email:     rpcResp.UserInfo.Email,
			Avatar:    rpcResp.UserInfo.Avatar,
			Status:    int(rpcResp.UserInfo.Status),
			IsAdmin:   rpcResp.UserInfo.IsAdmin,
			CreatedAt: rpcResp.UserInfo.CreatedAt,
		},
	}, nil
}
