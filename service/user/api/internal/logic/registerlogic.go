package logic

import (
	errutil "SChill/common/error"
	"SChill/service/user/api/internal/svc"
	"SChill/service/user/api/internal/types"
	"SChill/service/user/rpc/pb"
	"context"
	"github.com/zeromicro/go-zero/core/logc"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RegisterLogic) Register(req *types.RegisterReq) (resp *types.RegisterResp, err error) {
	// 1. 参数校验
	if err := req.Validate(); err != nil {
		logc.Infof(l.ctx, "注册参数错误: %v", err)
		return &types.RegisterResp{
			Code: errutil.ErrUsernameOrPasswordEmpty, // 用户名和密码参数
			Msg:  errutil.GetCodeMessage(errutil.ErrUsernameOrPasswordEmpty),
		}, nil
	}

	// 2. 调用 RPC 注册服务
	rpcResp, err := l.svcCtx.UserRpc.Register(l.ctx, &pb.RegisterReq{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		logc.Errorf(l.ctx, "用户服务: 注册 RPC 调用失败: %v", err)
		code, msg := errutil.ParseRpcError(err)
		return &types.RegisterResp{
			Code: code,
			Msg:  msg,
		}, nil
	}

	// 3. 成功响应
	return &types.RegisterResp{
		Code:   errutil.Success,
		Msg:    errutil.GetCodeMessage(errutil.Success),
		UserId: rpcResp.UserId,
	}, nil
}
