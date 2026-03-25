package logic

import (
	errutil "SChill/common/error"
	"SChill/common/jwt"
	"SChill/service/user/api/internal/svc"
	"SChill/service/user/api/internal/types"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshLogic {
	return &RefreshLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshLogic) Refresh(req *types.RefreshReq) (resp *types.RefreshResp, err error) {
	// 1. 参数校验
	if err := req.Validate(); err != nil {
		l.Infof("刷新 token 参数校验失败: %v", err)
		return &types.RefreshResp{
			Code: errutil.ErrInvalidParams,
			Msg:  errutil.GetCodeMessage(errutil.ErrInvalidParams),
		}, nil
	}

	// 2. 解析 refresh token 获取用户 ID
	userId, err := jwt.ParseRefreshToken(req.RefreshToken, l.svcCtx.Config.Jwt.RefreshSecret)
	if err != nil {
		l.Infof("refresh token 无效: %v", err)
		return &types.RefreshResp{
			Code: errutil.ErrInvalidRefreshToken,
			Msg:  errutil.GetCodeMessage(errutil.ErrInvalidRefreshToken),
		}, nil
	}

	// 3. 生成新的 access token 和 refresh token
	newAccessToken, err := jwt.GenerateAccessToken(
		l.svcCtx.Config.Jwt.AccessExpire,
		l.svcCtx.Config.Jwt.AccessSecret,
		userId,
	)
	if err != nil {
		logx.Errorf("生成 access token 失败: %v", err)
		return &types.RefreshResp{
			Code: errutil.ErrInternalError,
			Msg:  errutil.GetCodeMessage(errutil.ErrInternalError),
		}, nil
	}

	newRefreshToken, err := jwt.GenerateRefreshToken(
		l.svcCtx.Config.Jwt.RefreshExpire,
		l.svcCtx.Config.Jwt.RefreshSecret,
		userId,
	)
	if err != nil {
		logx.Errorf("生成 refresh token 失败: %v", err)
		return &types.RefreshResp{
			Code: errutil.ErrInternalError,
			Msg:  errutil.GetCodeMessage(errutil.ErrInternalError),
		}, nil
	}

	// 4. 成功返回
	return &types.RefreshResp{
		Code:            errutil.Success,
		Msg:             errutil.GetCodeMessage(errutil.Success),
		AccessToken:     newAccessToken,
		AccessExpireIn:  l.svcCtx.Config.Jwt.AccessExpire,
		RefreshToken:    newRefreshToken,
		RefreshExpireIn: l.svcCtx.Config.Jwt.RefreshExpire,
	}, nil
}
