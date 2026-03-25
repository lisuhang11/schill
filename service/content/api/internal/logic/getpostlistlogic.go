package logic

import (
	"context"

	errutil "SChill/common/error"
	"SChill/service/content/api/internal/svc"
	"SChill/service/content/api/internal/types"
	"SChill/service/content/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPostListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPostListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPostListLogic {
	return &GetPostListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPostListLogic) GetPostList(req *types.GetPostListReq) (resp *types.PostListResp, err error) {
	if err := req.Validate(); err != nil {
		logx.Errorf("获取帖子列表参数校验失败: %v", err)
		return &types.PostListResp{
			Code: errutil.ErrInvalidParams,
			Msg:  errutil.GetCodeMessage(errutil.ErrInvalidParams),
		}, nil
	}

	rpcResp, err := l.svcCtx.ContentRpc.GetPostList(l.ctx, &pb.GetPostListReq{
		UserId:   req.UserId,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		logx.Errorf("调用 RPC GetPostList 失败: %v", err)
		code, msg := errutil.ParseRpcError(err)
		return &types.PostListResp{
			Code: code,
			Msg:  msg,
		}, nil
	}

	var list []types.PostInfo
	for _, p := range rpcResp.List {
		list = append(list, types.PostInfo{
			Id:              p.Id,
			UserId:          p.UserId,
			Title:           p.Title,
			Cover:           p.Cover,
			CommentCount:    p.CommentCount,
			CollectionCount: p.CollectionCount,
			UpvoteCount:     p.UpvoteCount,
			ShareCount:      p.ShareCount,
			Visibility:      p.Visibility,
			IsTop:           p.IsTop,
			IsEssence:       p.IsEssence,
			IsLock:          p.IsLock,
			LatestRepliedAt: p.LatestRepliedAt,
			Tags:            p.Tags,
			CreatedAt:       p.CreatedAt,
			UpdatedAt:       p.UpdatedAt,
		})
	}

	return &types.PostListResp{
		Code:  errutil.Success,
		Msg:   errutil.GetCodeMessage(errutil.Success),
		Total: rpcResp.Total,
		List:  list,
	}, nil
}
