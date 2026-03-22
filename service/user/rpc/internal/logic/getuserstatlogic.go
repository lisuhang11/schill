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

type GetUserStatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserStatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserStatLogic {
	return &GetUserStatLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserStatLogic) GetUserStat(in *pb.GetUserStatReq) (*pb.GetUserStatResp, error) {
	var stat model.UserStat
	err := l.svcCtx.DB.WithContext(l.ctx).Where("user_id = ?", in.UserId).First(&stat).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, status.Error(codes.NotFound, "user stat not found")
		}
		logx.Errorf("get user stat failed: %v", err)
		return nil, status.Error(codes.Internal, "database error")
	}

	return &pb.GetUserStatResp{
		Stat: &pb.UserStat{
			UserId:          stat.UserID,
			PostCount:       uint64(stat.PostCount),
			CommentCount:    uint64(stat.CommentCount),
			FollowerCount:   uint64(stat.FollowerCount),
			FollowingCount:  uint64(stat.FollowingCount),
			LikeCount:       uint64(stat.LikeCount),
			CollectionCount: uint64(stat.CollectionCount),
			LastActiveTime:  stat.LastActiveTime,
		},
	}, nil
}
