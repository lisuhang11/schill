package logic

import (
	"context"

	"SChill/service/interaction/rpc/internal/model"
	"SChill/service/interaction/rpc/internal/svc"
	"SChill/service/interaction/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SharePostLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSharePostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SharePostLogic {
	return &SharePostLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

// 动态分享
func (l *SharePostLogic) SharePost(in *pb.SharePostReq) (*pb.SharePostResp, error) {
	share := &model.PostShare{
		PostID: in.PostId,
		UserID: in.UserId,
	}
	if err := l.svcCtx.DB.Create(share).Error; err != nil {
		logx.Errorf("创建分享记录失败: %v", err)
		return nil, err
	}

	var count int64
	err := l.svcCtx.DB.Model(&model.PostShare{}).Where("post_id = ?", in.PostId).Count(&count).Error
	if err != nil {
		logx.Errorf("获取分享数失败: %v", err)
	}

	return &pb.SharePostResp{
		Success:    true,
		ShareCount: count,
	}, nil
}
