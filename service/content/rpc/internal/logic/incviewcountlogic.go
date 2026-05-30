package logic

import (
	"context"

	errutil "SChill/common/error"
	"SChill/service/content/rpc/internal/model"
	"SChill/service/content/rpc/internal/svc"
	"SChill/service/content/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type IncViewCountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIncViewCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IncViewCountLogic {
	return &IncViewCountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *IncViewCountLogic) IncViewCount(in *pb.IncViewCountReq) (*pb.IncViewCountResp, error) {
	if in.GetPostId() == 0 {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}

	var post model.Post
	if err := l.svcCtx.DB.WithContext(l.ctx).
		Select("id").
		Where("id = ? AND deleted_at IS NULL", in.GetPostId()).
		Take(&post).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errutil.RpcBusinessError(errutil.ErrPostNotExist)
		}
		logx.Errorf("check post before inc view count failed: postId=%d err=%v", in.GetPostId(), err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	return &pb.IncViewCountResp{Success: true}, nil
}
