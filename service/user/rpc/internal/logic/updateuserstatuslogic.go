package logic

import (
	errutil "SChill/common/error"
	"SChill/service/user/rpc/internal/model"
	"context"

	"SChill/service/user/rpc/internal/svc"
	"SChill/service/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdateUserStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserStatusLogic {
	return &UpdateUserStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateUserStatusLogic) UpdateUserStatus(in *pb.UpdateUserStatusReq) (*pb.UpdateUserStatusResp, error) {
	if in.UserId == 0 {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}
	if in.Status < 1 || in.Status > 3 {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}

	var user model.User
	err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", in.UserId).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errutil.RpcBusinessError(errutil.ErrUserNotExist)
		}
		logx.Errorf("query user failed: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}
	if user.DeletedAt.Valid {
		return nil, errutil.RpcBusinessError(errutil.ErrUserNotExist)
	}

	result := l.svcCtx.DB.WithContext(l.ctx).Model(&model.User{}).
		Where("id = ?", in.UserId).
		Update("status", in.Status)

	if result.Error != nil {
		logx.Errorf("update user status failed: %v", result.Error)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	if result.RowsAffected == 0 {
		return nil, errutil.RpcBusinessError(errutil.ErrUserNotExist)
	}

	return &pb.UpdateUserStatusResp{Success: true}, nil
}
