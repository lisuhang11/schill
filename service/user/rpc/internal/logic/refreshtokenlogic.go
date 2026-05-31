package logic

import (
	"context"

	errutil "SChill/common/error"
	"SChill/common/jwt"
	"SChill/service/user/rpc/internal/model"
	"SChill/service/user/rpc/internal/svc"
	"SChill/service/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type RefreshTokenLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RefreshTokenLogic) RefreshToken(in *pb.RefreshTokenReq) (*pb.RefreshTokenResp, error) {
	if in.RefreshToken == "" {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}

	claims, err := jwt.ParseRefreshToken(in.RefreshToken, l.svcCtx.Config.Jwt.RefreshSecret)
	if err != nil {
		l.Infof("invalid refresh token: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidRefreshToken)
	}

	userID := claims.UserId

	var user model.User
	err = l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", userID).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errutil.RpcBusinessError(errutil.ErrInvalidRefreshToken)
		}
		logx.Errorf("find user for refresh token failed: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}
	if user.DeletedAt.Valid || user.Status != 1 {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidRefreshToken)
	}

	// Verify token version: if token has a version, it must match the current version in Redis.
	// This provides a revocation mechanism — incrementing the version invalidates all existing tokens.
	if claims.TokenVersion > 0 {
		currentVersion, err := getUserTokenVersion(l.ctx, l.svcCtx, userID)
		if err != nil {
			logx.Errorf("get user token version failed: userId=%d err=%v", userID, err)
		} else if claims.TokenVersion != currentVersion {
			l.Infof("token version mismatch: userId=%d tokenVersion=%d currentVersion=%d", userID, claims.TokenVersion, currentVersion)
			return nil, errutil.RpcBusinessError(errutil.ErrInvalidRefreshToken)
		}
	}

	tokenVersion := claims.TokenVersion
	if tokenVersion == 0 {
		tokenVersion = user.CreatedAt.Unix()
	}

	accessToken, err := jwt.GenerateAccessTokenWithVersion(
		l.svcCtx.Config.Jwt.AccessExpire,
		l.svcCtx.Config.Jwt.AccessSecret,
		userID,
		tokenVersion,
	)
	if err != nil {
		logx.Errorf("generate access token failed: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	refreshToken, err := jwt.GenerateRefreshTokenWithVersion(
		l.svcCtx.Config.Jwt.RefreshExpire,
		l.svcCtx.Config.Jwt.RefreshSecret,
		userID,
		tokenVersion,
	)
	if err != nil {
		logx.Errorf("generate refresh token failed: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	return &pb.RefreshTokenResp{
		UserId:          userID,
		AccessToken:     accessToken,
		AccessExpireIn:  l.svcCtx.Config.Jwt.AccessExpire,
		RefreshToken:    refreshToken,
		RefreshExpireIn: l.svcCtx.Config.Jwt.RefreshExpire,
	}, nil
}
