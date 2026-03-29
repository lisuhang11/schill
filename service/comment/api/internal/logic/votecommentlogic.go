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

type VoteCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVoteCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VoteCommentLogic {
	return &VoteCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VoteCommentLogic) VoteComment(req *types.VoteCommentReq) (resp *types.VoteCommentResp, err error) {
	userId, err := l.getUserIdFromContext()
	if err != nil || userId == 0 {
		return &types.VoteCommentResp{
			Code: errutil.ErrUnauthorized,
			Msg:  errutil.GetCodeMessage(errutil.ErrUnauthorized),
		}, nil
	}

	rpcResp, err := l.svcCtx.CommentRpc.VoteComment(l.ctx, &commentcenter.VoteCommentReq{
		CommentId: req.CommentId,
		UserId:    userId,
		VoteType:  req.VoteType,
	})
	if err != nil {
		logx.Errorf("调用 RPC VoteComment 失败: %v", err)
		code, msg := errutil.ParseRpcError(err)
		return &types.VoteCommentResp{
			Code: code,
			Msg:  msg,
		}, nil
	}

	return &types.VoteCommentResp{
		Code:         errutil.Success,
		Msg:          errutil.GetCodeMessage(errutil.Success),
		LikeCount:    rpcResp.LikeCount,
		DislikeCount: rpcResp.DislikeCount,
		IsLiked:      rpcResp.IsLiked,
		IsDisliked:   rpcResp.IsDisliked,
	}, nil
}

func (l *VoteCommentLogic) getUserIdFromContext() (uint64, error) {
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
