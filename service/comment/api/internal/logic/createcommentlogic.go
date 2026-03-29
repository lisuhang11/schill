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

type CreateCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCommentLogic {
	return &CreateCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateCommentLogic) CreateComment(req *types.CreateCommentReq) (resp *types.CreateCommentResp, err error) {
	if err := req.Validate(); err != nil {
		logx.Errorf("创建评论参数校验失败: %v", err)
		return &types.CreateCommentResp{
			Code: errutil.ErrInvalidParams,
			Msg:  errutil.GetCodeMessage(errutil.ErrInvalidParams),
		}, nil
	}

	userId, err := l.getUserIdFromContext()
	if err != nil || userId == 0 {
		return &types.CreateCommentResp{
			Code: errutil.ErrUnauthorized,
			Msg:  errutil.GetCodeMessage(errutil.ErrUnauthorized),
		}, nil
	}

	rpcResp, err := l.svcCtx.CommentRpc.CreateComment(l.ctx, &commentcenter.CreateCommentReq{
		UserId:        userId,
		PostId:        req.PostId,
		ParentId:      req.ParentId,
		ReplyToUserId: req.ReplyToUserId,
		Content:       req.Content,
	})
	if err != nil {
		logx.Errorf("调用 RPC CreateComment 失败: %v", err)
		code, msg := errutil.ParseRpcError(err)
		return &types.CreateCommentResp{
			Code: code,
			Msg:  msg,
		}, nil
	}

	if rpcResp == nil || rpcResp.Comment == nil {
		logx.Errorf("RPC 返回空响应")
		return &types.CreateCommentResp{
			Code: errutil.ErrInternalError,
			Msg:  errutil.GetCodeMessage(errutil.ErrInternalError),
		}, nil
	}

	commentInfo := types.CommentInfo{
		Id:              rpcResp.Comment.Id,
		PostId:          rpcResp.Comment.PostId,
		UserId:          rpcResp.Comment.UserId,
		ParentId:        rpcResp.Comment.ParentId,
		ReplyToUserId:   rpcResp.Comment.ReplyToUserId,
		Content:         rpcResp.Comment.Content,
		Level:           rpcResp.Comment.Level,
		ReplyCount:      rpcResp.Comment.ReplyCount,
		LikeCount:       rpcResp.Comment.LikeCount,
		CreatedAt:       rpcResp.Comment.CreatedAt,
		Username:        "",
		Avatar:          "",
		ReplyToUsername: "",
		IsLiked:         false,
	}

	return &types.CreateCommentResp{
		Code:    errutil.Success,
		Msg:     errutil.GetCodeMessage(errutil.Success),
		Comment: commentInfo,
	}, nil
}

func (l *CreateCommentLogic) getUserIdFromContext() (uint64, error) {
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
