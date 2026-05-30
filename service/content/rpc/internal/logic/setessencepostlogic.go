package logic

import (
	"context"
	"time"

	errutil "SChill/common/error"
	"SChill/service/content/rpc/internal/svc"
	"SChill/service/content/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetEssencePostLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetEssencePostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetEssencePostLogic {
	return &SetEssencePostLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetEssencePostLogic) SetEssencePost(in *pb.SetEssencePostReq) (*pb.SetEssencePostResp, error) {
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
			"is_essence": in.IsEssence,
			"updated_at": time.Now(),
		}).Error; err != nil {
		logx.Errorf("set essence post failed: postId=%d err=%v", in.PostId, err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	post.IsEssence = in.IsEssence
	post.UpdatedAt = time.Now()
	invalidatePostCachesByModel(l.ctx, l.svcCtx, post)
	publishContentChangedEvent(l.ctx, l.svcCtx, post, "updated", "", nil)

	return &pb.SetEssencePostResp{Success: true}, nil
}
