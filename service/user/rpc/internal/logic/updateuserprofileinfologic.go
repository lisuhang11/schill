package logic

import (
	"context"
	"time"

	"SChill/common/authctx"
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
	if in.UserProfile == nil {
		logx.Errorf("invalid update user profile request: missing user profile or user id")
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}

	// Resolve userId: prefer authenticated identity from gRPC interceptor,
	// fall back to request field for backward compatibility with callers
	// that do not yet set gRPC metadata (e.g. tests, internal scripts).
	authUserId := authctx.OptionalUserID(l.ctx)
	userId := in.UserId
	if userId == 0 {
		userId = in.UserProfile.UserId
	}
	if userId == 0 {
		logx.Errorf("invalid update user profile request: missing authenticated user id")
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}
	if authUserId != 0 && userId != authUserId {
		logx.Errorf("update user profile rejected: request userId=%d does not match authenticated userId=%d", userId, authUserId)
		return nil, errutil.RpcBusinessError(errutil.ErrNoPermission)
	}

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

	if in.UserProfile.Gender != nil {
		updates["gender"] = *in.UserProfile.Gender
	}
	if in.UserProfile.Birthday != nil {
		if *in.UserProfile.Birthday == "" {
			updates["birthday"] = nil
		} else {
			t, err := time.Parse("2006-01-02", *in.UserProfile.Birthday)
			if err != nil {
				logx.Errorf("invalid birthday format: %v", err)
				return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
			}
			updates["birthday"] = &t
		}
	}
	if in.UserProfile.Signature != nil {
		updates["signature"] = *in.UserProfile.Signature
	}
	if in.UserProfile.Location != nil {
		updates["location"] = *in.UserProfile.Location
	}
	if in.UserProfile.Website != nil {
		updates["website"] = *in.UserProfile.Website
	}
	if in.UserProfile.Company != nil {
		updates["company"] = *in.UserProfile.Company
	}
	if in.UserProfile.JobTitle != nil {
		updates["job_title"] = *in.UserProfile.JobTitle
	}
	if in.UserProfile.Education != nil {
		updates["education"] = *in.UserProfile.Education
	}

	if len(updates) > 0 {
		if err := l.svcCtx.DB.WithContext(l.ctx).Model(&profile).Updates(updates).Error; err != nil {
			logx.Errorf("update user profile failed: %v", err)
			return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
		}
		invalidateUserCaches(l.ctx, l.svcCtx, userId)
	}

	var updatedProfile model.UserProfile
	if err := l.svcCtx.DB.WithContext(l.ctx).Where("user_id = ?", userId).First(&updatedProfile).Error; err != nil {
		logx.Errorf("fetch updated profile failed: %v", err)
		updatedProfile = profile
	}

	pbProfile := &pb.UserProfile{
		UserId: userId,
	}
	gender := int64(updatedProfile.Gender)
	pbProfile.Gender = &gender
	signature := updatedProfile.Signature
	pbProfile.Signature = &signature
	location := updatedProfile.Location
	pbProfile.Location = &location
	website := updatedProfile.Website
	pbProfile.Website = &website
	company := updatedProfile.Company
	pbProfile.Company = &company
	jobTitle := updatedProfile.JobTitle
	pbProfile.JobTitle = &jobTitle
	education := updatedProfile.Education
	pbProfile.Education = &education
	if updatedProfile.Birthday != nil {
		birthday := updatedProfile.Birthday.Format("2006-01-02")
		pbProfile.Birthday = &birthday
	}

	return &pb.UpdateUserProfileInfoResp{
		UserProfile: pbProfile,
	}, nil
}
