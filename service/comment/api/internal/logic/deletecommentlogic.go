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

type DeleteCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCommentLogic {
	return &DeleteCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteCommentLogic) DeleteComment(req *types.DeleteCommentReq) (resp *types.DeleteCommentResp, err error) {
	userId, err := l.getUserIdFromContext()
	if err != nil || userId == 0 {
		return &types.DeleteCommentResp{
			Code: errutil.ErrUnauthorized,
			Msg:  errutil.GetCodeMessage(errutil.ErrUnauthorized),
		}, nil
	}

	_, err = l.svcCtx.CommentRpc.DeleteComment(l.ctx, &commentcenter.DeleteCommentReq{
		CommentId: req.CommentId,
		UserId:    userId,
	})
	if err != nil {
		logx.Errorf("调用 RPC DeleteComment 失败: %v", err)
		code, msg := errutil.ParseRpcError(err)
		return &types.DeleteCommentResp{
			Code: code,
			Msg:  msg,
		}, nil
	}

	return &types.DeleteCommentResp{
		Code: errutil.Success,
		Msg:  errutil.GetCodeMessage(errutil.Success),
	}, nil
}

func (l *DeleteCommentLogic) getUserIdFromContext() (uint64, error) {
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
