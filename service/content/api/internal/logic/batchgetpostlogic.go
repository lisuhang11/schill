// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	errutil "SChill/common/error"
	"SChill/service/content/api/internal/svc"
	"SChill/service/content/api/internal/types"
	"SChill/service/content/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type BatchGetPostLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBatchGetPostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchGetPostLogic {
	return &BatchGetPostLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchGetPostLogic) BatchGetPost(req *types.BatchGetPostReq) (resp *types.BatchGetPostResp, err error) {
	if err := req.Validate(); err != nil {
		logx.Errorf("批量获取帖子参数校验失败: %v", err)
		return &types.BatchGetPostResp{
			Code: errutil.ErrInvalidParams,
			Msg:  errutil.GetCodeMessage(errutil.ErrInvalidParams),
		}, nil
	}

	rpcResp, err := l.svcCtx.ContentRpc.BatchGetPost(l.ctx, &pb.BatchGetPostReq{
		PostIds: req.PostIds,
	})
	if err != nil {
		logx.Errorf("调用 RPC BatchGetPost 失败: %v", err)
		code, msg := errutil.ParseRpcError(err)
		return &types.BatchGetPostResp{
			Code: code,
			Msg:  msg,
		}, nil
	}

	var posts []types.PostInfo
	for _, p := range rpcResp.Posts {
		posts = append(posts, types.PostInfo{
			Id:           p.Id,
			UserId:       p.UserId,
			Title:        p.Title,
			Cover:        p.Cover,
			Type:         p.Type,
			Status:       p.Status,
			ViewCount:    p.ViewCount,
			LikeCount:    p.LikeCount,
			CommentCount: p.CommentCount,
			CreatedAt:    p.CreatedAt,
			UpdatedAt:    p.UpdatedAt,
		})
	}

	return &types.BatchGetPostResp{
		Code:  errutil.Success,
		Msg:   errutil.GetCodeMessage(errutil.Success),
		Posts: posts,
	}, nil
}
