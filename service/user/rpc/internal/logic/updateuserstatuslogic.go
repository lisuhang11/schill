package logic

import (
	"SChill/service/user/rpc/internal/model"
	"context"

	"SChill/service/user/rpc/internal/svc"
	"SChill/service/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	// 参数校验
	if in.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}
	// 状态值必须是 1、2 或 3
	if in.Status < 1 || in.Status > 3 {
		return nil, status.Error(codes.InvalidArgument, "invalid status, must be 1, 2 or 3")
	}

	// 检查用户是否存在且未被软删除
	var user model.User
	err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", in.UserId).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		logx.Errorf("query user failed: %v", err)
		return nil, status.Error(codes.Internal, "database error")
	}
	if user.DeletedAt.Valid {
		return nil, status.Error(codes.NotFound, "user deleted")
	}

	// 更新状态
	result := l.svcCtx.DB.WithContext(l.ctx).Model(&model.User{}).
		Where("id = ?", in.UserId).
		Update("status", in.Status)

	if result.Error != nil {
		logx.Errorf("update user status failed: %v", result.Error)
		return nil, status.Error(codes.Internal, "update failed")
	}

	if result.RowsAffected == 0 {
		return nil, status.Error(codes.NotFound, "user not found")
	}

	return &pb.UpdateUserStatusResp{Success: true}, nil
}
