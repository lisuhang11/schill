package logic

import (
	"SChill/common/errorcode"
	"SChill/common/jwt"
	"SChill/service/user/api/internal/svc"
	"SChill/service/user/api/internal/types"
	"SChill/service/user/rpc/pb"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *LoginLogic) Login(req *types.LoginReq) (resp *types.LoginResp, err error) {
	// 1. 参数校验
	if err := req.Validate(); err != nil {
		// 记录详细的验证错误
		l.Infof("登录参数验证失败: %v", err)
		// 返回通用的参数错误码
		return &types.LoginResp{
			Code: errorcode.ErrUsernameOrPasswordEmpty, // 用户名或密码格式
			Msg:  err.Error(),
		}, nil
	}

	// 2. 调用 RPC 登录
	rpcResp, err := l.svcCtx.UserRpc.Login(l.ctx, &pb.LoginReq{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		logx.Errorf("登录 RPC 调用失败: %v", err)
		st, _ := status.FromError(err)
		switch st.Code() {
		case codes.NotFound, codes.Unauthenticated:
			// 统一返回“用户名或密码错误”
			return &types.LoginResp{
				Code: errorcode.ErrInvalidCredentials,
				Msg:  errorcode.GetCodeMessage(errorcode.ErrInvalidCredentials),
			}, nil
		case codes.PermissionDenied:
			return &types.LoginResp{
				Code: errorcode.ErrAccountAbnormal,
				Msg:  errorcode.GetCodeMessage(errorcode.ErrAccountAbnormal),
			}, nil
		default:
			return &types.LoginResp{
				Code: errorcode.ErrInternalError,
				Msg:  errorcode.GetCodeMessage(errorcode.ErrInternalError),
			}, nil
		}
	}

	// 3. 生成 Access Token 和 Refresh Token
	accessToken, err := jwt.GenerateAccessToken(
		l.svcCtx.Config.Jwt.AccessExpire,
		l.svcCtx.Config.Jwt.AccessSecret,
		rpcResp.UserId,
	)
	if err != nil {
		logx.Errorf("生成 AccessToken 失败: %v", err)
		return &types.LoginResp{
			Code: errorcode.ErrInternalError,
			Msg:  errorcode.GetCodeMessage(errorcode.ErrInternalError),
		}, nil
	}

	refreshToken, err := jwt.GenerateRefreshToken(
		l.svcCtx.Config.Jwt.RefreshExpire,
		l.svcCtx.Config.Jwt.RefreshSecret,
		rpcResp.UserId,
	)
	if err != nil {
		logx.Errorf("生成 RefreshToken 失败: %v", err)
		return &types.LoginResp{
			Code: errorcode.ErrInternalError,
			Msg:  errorcode.GetCodeMessage(errorcode.ErrInternalError),
		}, nil
	}

	// 4. 成功响应
	return &types.LoginResp{
		Code:            errorcode.Success,
		Msg:             errorcode.GetCodeMessage(errorcode.Success),
		UserId:          rpcResp.UserId,
		AccessToken:     accessToken,
		AccessExpireIn:  l.svcCtx.Config.Jwt.AccessExpire,
		RefreshToken:    refreshToken,
		RefreshExpireIn: l.svcCtx.Config.Jwt.RefreshExpire,
	}, nil
}
