package logic

import (
	"context"
	"time"

	"SChill/service/user/rpc/internal/model"
	"SChill/service/user/rpc/internal/svc"
	"SChill/service/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type UpdateUserProfileInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdateUserProfileInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdateUserProfileInfoLogic {
	return &UpdateUserProfileInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdateUserProfileInfoLogic) UpdateUserProfileInfo(in *pb.UpdateUserProfileInfoReq) (*pb.UpdateUserProfileInfoResp, error) {
	if in.UserProfile == nil || in.UserProfile.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid user profile")
	}
	userId := in.UserProfile.UserId

	// 先检查用户扩展信息是否存在
	var profile model.UserProfile
	err := l.svcCtx.DB.WithContext(l.ctx).Where("user_id = ?", userId).First(&profile).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "user profile not found")
		}
		logx.Errorf("find user profile failed: %v", err)
		return nil, status.Error(codes.Internal, "database error")
	}

	// 构建更新字段
	updates := make(map[string]interface{})

	if in.UserProfile.Gender == 1 || in.UserProfile.Gender == 2 {
		updates["gender"] = in.UserProfile.Gender
	}
	if in.UserProfile.Birthday != "" {
		t, err := time.Parse("2006-01-02", in.UserProfile.Birthday)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid birthday format: %v", err)
		}
		updates["birthday"] = &t
	}
	// 以下字段允许更新为空字符串
	updates["signature"] = in.UserProfile.Signature
	updates["location"] = in.UserProfile.Location
	updates["website"] = in.UserProfile.Website
	updates["company"] = in.UserProfile.Company
	updates["job_title"] = in.UserProfile.JobTitle
	updates["education"] = in.UserProfile.Education

	// 执行更新
	if len(updates) > 0 {
		if err := l.svcCtx.DB.WithContext(l.ctx).Model(&profile).Updates(updates).Error; err != nil {
			logx.Errorf("update user profile failed: %v", err)
			return nil, status.Error(codes.Internal, "update failed")
		}
	}

	// 查询更新后的数据
	var updatedProfile model.UserProfile
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("user_id = ?", userId).First(&updatedProfile).Error; err != nil {
		logx.Errorf("fetch updated profile failed: %v", err)
		updatedProfile = profile // 降级使用旧数据
	}

	// 转换为 pb 返回
	pbProfile := &pb.UserProfile{
		UserId:    updatedProfile.UserID,
		Gender:    int64(updatedProfile.Gender),
		Signature: updatedProfile.Signature,
		Location:  updatedProfile.Location,
		Website:   updatedProfile.Website,
		Company:   updatedProfile.Company,
		JobTitle:  updatedProfile.JobTitle,
		Education: updatedProfile.Education,
	}
	if updatedProfile.Birthday != nil {
		pbProfile.Birthday = updatedProfile.Birthday.Format("2006-01-02")
	}

	return &pb.UpdateUserProfileInfoResp{
		UserProfile: pbProfile,
	}, nil
}
