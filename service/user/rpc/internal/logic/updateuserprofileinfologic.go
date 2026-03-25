package logic

import (
	"context"
	"time"

	errutil "SChill/common/error"
	"SChill/service/user/rpc/internal/model"
	"SChill/service/user/rpc/internal/svc"
	"SChill/service/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
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
	logx.Infof("RPC收到更新用户资料请求: userId=[%d], birthday=[%s], gender=[%d]",
		in.UserProfile.UserId, in.UserProfile.Birthday, in.UserProfile.Gender)

	if in.UserProfile == nil || in.UserProfile.UserId == 0 {
		logx.Errorf("参数错误: UserProfile或UserId为空")
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}
	userId := in.UserProfile.UserId

	var profile model.UserProfile
	err := l.svcCtx.DB.WithContext(l.ctx).Where("user_id = ?", userId).First(&profile).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errutil.RpcBusinessError(errutil.ErrUserNotExist)
		}
		logx.Errorf("find user profile failed: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	updates := make(map[string]interface{})

	if in.UserProfile.Gender == 1 || in.UserProfile.Gender == 2 {
		updates["gender"] = in.UserProfile.Gender
	}
	if in.UserProfile.Birthday != "" {
		logx.Infof("解析生日: [%s]", in.UserProfile.Birthday)
		t, err := time.Parse("2006-01-02", in.UserProfile.Birthday)
		if err != nil {
			logx.Errorf("生日格式错误: %v", err)
			return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
		}
		updates["birthday"] = &t
	}
	updates["signature"] = in.UserProfile.Signature
	updates["location"] = in.UserProfile.Location
	updates["website"] = in.UserProfile.Website
	updates["company"] = in.UserProfile.Company
	updates["job_title"] = in.UserProfile.JobTitle
	updates["education"] = in.UserProfile.Education

	if len(updates) > 0 {
		if err := l.svcCtx.DB.WithContext(l.ctx).Model(&profile).Updates(updates).Error; err != nil {
			logx.Errorf("update user profile failed: %v", err)
			return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
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
