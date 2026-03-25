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

	var profile model.UserProfile
	err := l.svcCtx.DB.WithContext(l.ctx).Where("user_id = ?", in.UserId).First(&profile).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errutil.RpcBusinessError(errutil.ErrUserNotExist)
		}
		logx.Errorf("find user profile failed: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	pbProfile := &pb.UserProfile{
		UserId:    profile.UserID,
		Gender:    int64(profile.Gender),
		Signature: profile.Signature,
		Location:  profile.Location,
		Website:   profile.Website,
		Company:   profile.Company,
		JobTitle:  profile.JobTitle,
		Education: profile.Education,
	}
	if profile.Birthday != nil {
		pbProfile.Birthday = profile.Birthday.Format("2006-01-02")
	}

	return &pb.GetUserProfileInfoResp{
		Profile: pbProfile,
	}, nil
}
