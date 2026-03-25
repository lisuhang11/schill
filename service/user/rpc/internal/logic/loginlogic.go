package logic

import (
	"context"
	"time"

	"SChill/common/cryptx"
	errutil "SChill/common/error"
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

	return &pb.LoginResp{
		UserId: user.ID,
	}, nil
}
