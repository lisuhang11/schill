package logic

import (
	"context"

	"SChill/service/user/rpc/internal/model"
	"SChill/service/user/rpc/internal/svc"
	"SChill/service/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type GetUserInfoLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserInfoLogic) GetUserInfo(in *pb.GetUserInfoReq) (*pb.GetUserInfoResp, error) {
	// 1. 查询用户基础信息
	var user model.User
	err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", in.UserId).First(&user).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		logx.Errorf("find user failed: %v", err)
		return nil, status.Error(codes.Internal, "database error")
	}
	if user.DeletedAt.Valid {
		return nil, status.Error(codes.NotFound, "user deleted")
	}

	// 2. 查询扩展信息（可能不存在）
	var profile model.UserProfile
	_ = l.svcCtx.DB.WithContext(l.ctx).Where("user_id = ?", in.UserId).First(&profile).Error

	// 3. 查询统计信息（可能不存在）
	var stat model.UserStat
	_ = l.svcCtx.DB.WithContext(l.ctx).Where("user_id = ?", in.UserId).First(&stat).Error

	// 组装响应
	userInfo := &pb.UserInfo{
		Id:        user.ID,
		Username:  user.Username,
		Phone:     getStringValue(user.Phone),
		Email:     getStringValue(user.Email),
		Avatar:    user.Avatar,
		Status:    int32(user.Status),
		IsAdmin:   user.IsAdmin == 1,
		CreatedAt: user.CreatedAt.Unix(),
		Nickname:  "", // 表中暂无 nickname，如需可添加字段
	}

	resp := &pb.GetUserInfoResp{
		UserInfo: userInfo,
	}

	// 如果 profile 存在，填充
	if profile.ID != 0 {
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
		resp.Profile = pbProfile
	}

	// 如果 stat 存在，填充
	if stat.ID != 0 {
		resp.Stat = &pb.UserStat{
			UserId:          stat.UserID,
			PostCount:       uint64(stat.PostCount),
			CommentCount:    uint64(stat.CommentCount),
			FollowerCount:   uint64(stat.FollowerCount),
			FollowingCount:  uint64(stat.FollowingCount),
			LikeCount:       uint64(stat.LikeCount),
			CollectionCount: uint64(stat.CollectionCount),
			LastActiveTime:  stat.LastActiveTime,
		}
	}

	return resp, nil
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
