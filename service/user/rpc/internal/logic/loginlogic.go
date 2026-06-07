package logic

import (
	"context"
	"time"

	"SChill/common/cryptx"
	errutil "SChill/common/error"
	"SChill/common/jwt"
	"SChill/service/user/rpc/internal/model"
	"SChill/service/user/rpc/internal/svc"
	"SChill/service/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *pb.LoginReq) (*pb.LoginResp, error) {
	var user model.User
	err := l.svcCtx.DB.WithContext(l.ctx).Where("username = ?", in.Username).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errutil.RpcBusinessError(errutil.ErrInvalidCredentials)
		}
		logx.Errorf("查询用户失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	if user.DeletedAt.Valid {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidCredentials)
	}

	if !cryptx.PasswordVerify(user.PasswordHash, in.Password) {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidCredentials)
	}

	if user.Status != 1 {
		return nil, errutil.RpcBusinessError(errutil.ErrAccountAbnormal)
	}

	now := time.Now()
	user.LastLoginTime = &now
	if err := l.svcCtx.DB.WithContext(l.ctx).Model(&user).Update("last_login_time", now).Error; err != nil {
		logx.Errorf("更新用户登录时间失败: %v", err)
	}

	// 更新用户最后活跃时间
	var userStat model.UserStat
	errStat := l.svcCtx.DB.WithContext(l.ctx).Where("user_id = ?", user.UserID).First(&userStat).Error
	if errStat != nil {
		if errStat == gorm.ErrRecordNotFound {
			// 用户统计记录不存在，创建新记录
			userStat = model.UserStat{
				UserID:         user.UserID,
				LastActiveTime: &now,
			}
			if errCreate := l.svcCtx.DB.WithContext(l.ctx).Create(&userStat).Error; errCreate != nil {
				logx.Errorf("创建用户统计记录失败: %v", errCreate)
			}
		} else {
			logx.Errorf("查询用户统计记录失败: %v", errStat)
		}
	} else {
		// 用户统计记录存在，更新最后活跃时间
		if errUpdate := l.svcCtx.DB.WithContext(l.ctx).Model(&userStat).Update("last_active_time", now).Error; errUpdate != nil {
			logx.Errorf("更新用户最后活跃时间失败: %v", errUpdate)
		}
	}

	invalidateUserCaches(l.ctx, l.svcCtx, user.UserID)

	tokenVersion, err := getUserTokenVersion(l.ctx, l.svcCtx, user.UserID)
	if err != nil {
		logx.Errorf("get user token version failed: %v", err)
	}
	if tokenVersion == 0 {
		tokenVersion = user.CreatedAt.Unix()
	}

	accessToken, err := jwt.GenerateAccessTokenWithVersion(
		l.svcCtx.Config.Jwt.AccessExpire,
		l.svcCtx.Config.Jwt.AccessSecret,
		user.UserID,
		tokenVersion,
	)
	if err != nil {
		logx.Errorf("generate access token failed: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	refreshToken, err := jwt.GenerateRefreshTokenWithVersion(
		l.svcCtx.Config.Jwt.RefreshExpire,
		l.svcCtx.Config.Jwt.RefreshSecret,
		user.UserID,
		tokenVersion,
	)
	if err != nil {
		logx.Errorf("generate refresh token failed: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	return &pb.LoginResp{
		UserId:          user.UserID,
		AccessToken:     accessToken,
		AccessExpireIn:  l.svcCtx.Config.Jwt.AccessExpire,
		RefreshToken:    refreshToken,
		RefreshExpireIn: l.svcCtx.Config.Jwt.RefreshExpire,
	}, nil
}
