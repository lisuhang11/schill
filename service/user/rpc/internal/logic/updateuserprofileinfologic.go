package logic

import (
	"context"
	"database/sql"
	"time"

	"SChill/service/user/rpc/internal/model"
	"SChill/service/user/rpc/internal/svc"
	"SChill/service/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
	// 参数校验
	if in.UserProfile == nil {
		return nil, status.Error(codes.InvalidArgument, "user profile is required")
	}
	userId := in.UserProfile.UserId
	if userId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	// 转换 birthday 字符串为 sql.NullTime
	var birthdaySql sql.NullTime
	if in.UserProfile.Birthday != "" {
		t, err := time.Parse("2006-01-02", in.UserProfile.Birthday)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid birthday format: %v", err)
		}
		birthdaySql = sql.NullTime{
			Time:  t,
			Valid: true,
		}
	} else {
		birthdaySql = sql.NullTime{Valid: false}
	}

	// 构造 model 对象
	profileModel := &model.UserProfile{
		UserId:    userId,
		Gender:    in.UserProfile.Gender,
		Birthday:  birthdaySql,
		Signature: in.UserProfile.Signature,
		Location:  in.UserProfile.Location,
		Website:   in.UserProfile.Website,
		Company:   in.UserProfile.Company,
		JobTitle:  in.UserProfile.JobTitle,
		Education: in.UserProfile.Education,
	}

	// 先检查是否存在，若不存在则返回
	_, err := l.svcCtx.UserProfileModel.FindOneByUserId(l.ctx, userId)
	if err == model.ErrNotFound {
		logx.Errorf("insert user profile failed: %v", err)
		return nil, status.Error(codes.Internal, "user not exists")
	} else if err != nil {
		logx.Errorf("find user profile failed: %v", err)
		return nil, status.Error(codes.Internal, "find user profile failed")
	} else {
		// 更新记录
		err = l.svcCtx.UserProfileModel.UpdateByUserId(l.ctx, profileModel)
		if err != nil {
			logx.Errorf("update user profile failed: %v", err)
			return nil, status.Error(codes.Internal, "update user profile failed")
		}
	}

	// 查询更新后的完整数据（可选，确保返回最新）
	updated, err := l.svcCtx.UserProfileModel.FindOneByUserId(l.ctx, userId)
	if err != nil {
		logx.Errorf("find updated user profile failed: %v", err)
		// 即使查询失败，仍返回请求的数据，但记录错误
	} else {
		profileModel = updated
	}

	// 转换为 pb 返回
	pbProfile := &pb.UserProfile{
		UserId:    profileModel.UserId,
		Gender:    profileModel.Gender,
		Birthday:  profileModel.Birthday.Time.Format("2006-01-02"),
		Signature: profileModel.Signature,
		Location:  profileModel.Location,
		Website:   profileModel.Website,
		Company:   profileModel.Company,
		JobTitle:  profileModel.JobTitle,
		Education: profileModel.Education,
	}
	if !profileModel.Birthday.Valid {
		pbProfile.Birthday = ""
	}

	return &pb.UpdateUserProfileInfoResp{
		UserProfile: pbProfile,
	}, nil
}
