package logic

import (
	"context"
	"time"

	errutil "SChill/common/error"
	"SChill/service/content/rpc/internal/svc"
	"SChill/service/content/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetTopPostLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetTopPostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetTopPostLogic {
	return &SetTopPostLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetTopPostLogic) SetTopPost(in *pb.SetTopPostReq) (*pb.SetTopPostResp, error) {
	if in.UserId == 0 || in.PostId == 0 {
		return nil, errutil.RpcBusinessError(errutil.ErrUnauthorized)
	}

	post, err := loadOwnedPost(l.ctx, l.svcCtx.DBWrite, in.PostId, in.UserId)
	if err != nil {
		return nil, err
	}

	if err := l.svcCtx.DBWrite.WithContext(l.ctx).
		Model(post).
		Updates(map[string]interface{}{
			"is_top":     in.IsTop,
			"updated_at": time.Now(),
		}).Error; err != nil {
		logx.Errorf("set top post failed: postId=%d err=%v", in.PostId, err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	post.IsTop = in.IsTop
	post.UpdatedAt = time.Now()
	invalidatePostCachesByModel(l.ctx, l.svcCtx, post)

	return &pb.SetTopPostResp{Success: true}, nil
}
