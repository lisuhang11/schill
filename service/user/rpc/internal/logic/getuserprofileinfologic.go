package logic

import (
	"context"

	errutil "SChill/common/error"
	"SChill/service/user/rpc/internal/model"
	"SChill/service/user/rpc/internal/svc"
	"SChill/service/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type GetUserProfileInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserProfileInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserProfileInfoLogic {
	return &GetUserProfileInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserProfileInfoLogic) GetUserProfileInfo(in *pb.GetUserProfileInfoReq) (*pb.GetUserProfileInfoResp, error) {
	if in.UserId == 0 {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}

	cacheKey := buildUserProfileCacheKey(in.UserId)

	var resp pb.GetUserProfileInfoResp
	err := l.svcCtx.Cache.TakeCtx(l.ctx, &resp, cacheKey, func(val interface{}) error {
		var profile model.UserProfile
		err := l.svcCtx.DB.WithContext(l.ctx).Where("user_id = ?", in.UserId).First(&profile).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil
			}
			logx.Errorf("find user profile failed: %v", err)
			return errutil.RpcBusinessError(errutil.ErrInternalError)
		}

		gender := int64(profile.Gender)
		signature := profile.Signature
		location := profile.Location
		website := profile.Website
		company := profile.Company
		jobTitle := profile.JobTitle
		education := profile.Education
		pbProfile := &pb.UserProfile{
			UserId:    profile.UserID,
			Gender:    &gender,
			Signature: &signature,
			Location:  &location,
			Website:   &website,
			Company:   &company,
			JobTitle:  &jobTitle,
			Education: &education,
		}
		if profile.Birthday != nil {
			birthday := profile.Birthday.Format("2006-01-02")
			pbProfile.Birthday = &birthday
		}

		*val.(*pb.GetUserProfileInfoResp) = pb.GetUserProfileInfoResp{
			Profile: pbProfile,
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	if resp.Profile == nil {
		return nil, errutil.RpcBusinessError(errutil.ErrUserNotExist)
	}

	return &resp, nil
}
