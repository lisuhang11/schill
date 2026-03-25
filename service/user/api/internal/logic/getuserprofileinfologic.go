package logic

import (
	errutil "SChill/common/error"
	"SChill/service/user/api/internal/svc"
	"SChill/service/user/api/internal/types"
	"SChill/service/user/rpc/pb"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserProfileInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserProfileInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserProfileInfoLogic {
	return &GetUserProfileInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserProfileInfoLogic) GetUserProfileInfo(req *types.GetUserProfileInfoReq) (resp *types.GetUserProfileInfoResp, err error) {
	rpcResp, err := l.svcCtx.UserRpc.GetUserProfileInfo(l.ctx, &pb.GetUserProfileInfoReq{
		UserId: req.UserId,
	})
	if err != nil {
		logx.Errorf("调用 RPC GetUserProfileInfo 失败: %v", err)
		code, msg := errutil.ParseRpcError(err)
		return &types.GetUserProfileInfoResp{
			Code: code,
			Msg:  msg,
		}, nil
	}

	profile := types.UserProfileInfo{
		UserId:    rpcResp.Profile.UserId,
		Gender:    int(rpcResp.Profile.Gender),
		Birthday:  rpcResp.Profile.Birthday,
		Signature: rpcResp.Profile.Signature,
		Location:  rpcResp.Profile.Location,
		Website:   rpcResp.Profile.Website,
		Company:   rpcResp.Profile.Company,
		JobTitle:  rpcResp.Profile.JobTitle,
		Education: rpcResp.Profile.Education,
	}

	return &types.GetUserProfileInfoResp{
		Code:    errutil.Success,
		Msg:     errutil.GetCodeMessage(errutil.Success),
		Profile: profile,
	}, nil
}
