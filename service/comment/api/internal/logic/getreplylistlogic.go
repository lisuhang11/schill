package logic

import (
	"context"
	"encoding/json"

	errutil "SChill/common/error"
	"SChill/service/comment/api/internal/svc"
	"SChill/service/comment/api/internal/types"
	"SChill/service/comment/rpc/commentcenter"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetReplyListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetReplyListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetReplyListLogic {
	return &GetReplyListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetReplyListLogic) GetReplyList(req *types.GetReplyListReq) (resp *types.CommentListResp, err error) {
	rpcResp, err := l.svcCtx.CommentRpc.GetReplyList(l.ctx, &commentcenter.GetReplyListReq{
		CommentId: req.CommentId,
		Cursor:    req.Cursor,
		PageSize:  req.PageSize,
	})
	if err != nil {
		logx.Errorf("调用 RPC GetReplyList 失败: %v", err)
		code, msg := errutil.ParseRpcError(err)
		return &types.CommentListResp{
			Code: code,
			Msg:  msg,
		}, nil
	}

	var list []types.CommentItem
	for _, rpcComment := range rpcResp.List {
		commentInfo := types.CommentInfo{
			Id:              rpcComment.Id,
			PostId:          rpcComment.PostId,
			UserId:          rpcComment.UserId,
			ParentId:        rpcComment.ParentId,
			ReplyToUserId:   rpcComment.ReplyToUserId,
			Content:         rpcComment.Content,
			Level:           rpcComment.Level,
			ReplyCount:      rpcComment.ReplyCount,
			LikeCount:       rpcComment.LikeCount,
			CreatedAt:       rpcComment.CreatedAt,
			Username:        "",
			Avatar:          "",
			ReplyToUsername: "",
			IsLiked:         false,
		}
		list = append(list, types.CommentItem{
			Root:           commentInfo,
			Replies:        []types.CommentInfo{},
			HasMoreReplies: false,
		})
	}

	return &types.CommentListResp{
		Code:  errutil.Success,
		Msg:   errutil.GetCodeMessage(errutil.Success),
		Total: rpcResp.Total,
		List:  list,
	}, nil
}

func (l *GetReplyListLogic) getUserIdFromContext() (uint64, error) {
	userIdVal := l.ctx.Value("userId")
	switch v := userIdVal.(type) {
	case uint64:
		return v, nil
	case int64:
		return uint64(v), nil
	case int:
		return uint64(v), nil
	case float64:
		return uint64(v), nil
	case json.Number:
		if userId, err := v.Int64(); err == nil {
			return uint64(userId), nil
		}
		if userId, err := v.Float64(); err == nil {
			return uint64(userId), nil
		}
	}
	return 0, nil
}
