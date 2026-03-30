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
	if in.PostId == 0 {
		return nil, errutil.RpcBusinessError(errutil.ErrPostNotExist)
	}

	var post model.Post
	err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", in.PostId).First(&post).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errutil.RpcBusinessError(errutil.ErrPostNotExist)
		}
		logx.Errorf("查询帖子失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	return &pb.IncViewCountResp{
		Success: true,
	}, nil
}
