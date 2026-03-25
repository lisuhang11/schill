package logic

import (
	"context"

	errutil "SChill/common/error"
	"SChill/service/content/api/internal/svc"
	"SChill/service/content/api/internal/types"
	"SChill/service/content/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPostDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPostDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPostDetailLogic {
	return &GetPostDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPostDetailLogic) GetPostDetail(req *types.GetPostDetailReq) (resp *types.PostDetailResp, err error) {
	if err := req.Validate(); err != nil {
		logx.Errorf("获取帖子详情参数校验失败: %v", err)
		return &types.PostDetailResp{
			Code: errutil.ErrInvalidParams,
			Msg:  errutil.GetCodeMessage(errutil.ErrInvalidParams),
		}, nil
	}

	rpcResp, err := l.svcCtx.ContentRpc.GetPostDetail(l.ctx, &pb.GetPostDetailReq{
		PostId: req.PostId,
	})
	if err != nil {
		logx.Errorf("调用 RPC GetPostDetail 失败: %v", err)
		code, msg := errutil.ParseRpcError(err)
		return &types.PostDetailResp{
			Code: code,
			Msg:  msg,
		}, nil
	}

	var topics []types.PostTopic
	for _, t := range rpcResp.Topics {
		topics = append(topics, types.PostTopic{
			PostId:    t.PostId,
			TopicId:   t.TopicId,
			TopicName: t.TopicName,
		})
	}

	var content string
	if len(rpcResp.Contents) > 0 {
		for _, c := range rpcResp.Contents {
			if c.Type == 2 {
				content = c.Content
				break
			}
		}
	}

	return &types.PostDetailResp{
		Code: errutil.Success,
		Msg:  errutil.GetCodeMessage(errutil.Success),
		Post: types.PostInfo{
			Id:              rpcResp.Post.Id,
			UserId:          rpcResp.Post.UserId,
			Title:           rpcResp.Post.Title,
			Cover:           rpcResp.Post.Cover,
			CommentCount:    rpcResp.Post.CommentCount,
			CollectionCount: rpcResp.Post.CollectionCount,
			UpvoteCount:     rpcResp.Post.UpvoteCount,
			ShareCount:      rpcResp.Post.ShareCount,
			Visibility:      rpcResp.Post.Visibility,
			IsTop:           rpcResp.Post.IsTop,
			IsEssence:       rpcResp.Post.IsEssence,
			IsLock:          rpcResp.Post.IsLock,
			LatestRepliedAt: rpcResp.Post.LatestRepliedAt,
			Tags:            rpcResp.Post.Tags,
			CreatedAt:       rpcResp.Post.CreatedAt,
			UpdatedAt:       rpcResp.Post.UpdatedAt,
		},
		Content: content,
		Topics:  topics,
	}, nil
}
