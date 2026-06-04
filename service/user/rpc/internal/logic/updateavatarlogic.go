package logic

import (
	"context"

	"SChill/common/authctx"
	errutil "SChill/common/error"
	"SChill/service/user/rpc/internal/model"
	"SChill/service/user/rpc/internal/svc"
	"SChill/service/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdateAvatarLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateAvatarLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateAvatarLogic {
	return &UpdateAvatarLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateAvatarLogic) UpdateAvatar(in *pb.UpdateAvatarReq) (*pb.UpdateAvatarResp, error) {
	// 参数校验
	if in.UserId == 0 {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}
	if in.AvatarUrl == "" {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}

	// Verify caller identity against authenticated user from gRPC interceptor.
	// Falls back to trusting the request field if no interceptor-provided userId
	// exists (backward compatibility with callers not yet setting metadata).
	if authUserId := authctx.OptionalUserID(l.ctx); authUserId != 0 && authUserId != in.UserId {
		logx.Errorf("update avatar rejected: request userId=%d does not match authenticated userId=%d", in.UserId, authUserId)
		return nil, errutil.RpcBusinessError(errutil.ErrNoPermission)
	}

	var user model.User
	err := l.svcCtx.DB.WithContext(l.ctx).Where("user_id = ?", in.UserId).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errutil.RpcBusinessError(errutil.ErrUserNotExist)
		}
		logx.Errorf("find user failed: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}
	if user.DeletedAt.Valid {
		return nil, errutil.RpcBusinessError(errutil.ErrUserNotExist)
	}

	// 保存新的头像url
	user.Avatar = in.AvatarUrl
	if err := l.svcCtx.DB.WithContext(l.ctx).Save(&user).Error; err != nil {
		logx.Errorf("update avatar failed: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}
	invalidateUserCaches(l.ctx, l.svcCtx, in.UserId)

	return &pb.UpdateAvatarResp{
		AvatarUrl: user.Avatar,
	}, nil
}
