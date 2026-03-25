// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"encoding/json"

	errutil "SChill/common/error"
	"SChill/service/content/api/internal/svc"
	"SChill/service/content/api/internal/types"
	"SChill/service/content/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreatePostLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreatePostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePostLogic {
	return &CreatePostLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreatePostLogic) CreatePost(req *types.CreatePostReq) (resp *types.CreatePostResp, err error) {
	if err := req.Validate(); err != nil {
		logx.Errorf("创建帖子参数校验失败: %v", err)
		return &types.CreatePostResp{
			Code: errutil.ErrInvalidParams,
			Msg:  errutil.GetCodeMessage(errutil.ErrInvalidParams),
		}, nil
	}

	userId, err := l.getUserIdFromContext()
	if err != nil || userId == 0 {
		return &types.CreatePostResp{
			Code: errutil.ErrUnauthorized,
			Msg:  errutil.GetCodeMessage(errutil.ErrUnauthorized),
		}, nil
	}

	contents := []*pb.PostContentItem{
		{
			Type:    2,
			Content: req.Content,
			Sort:    10,
		},
	}

	rpcResp, err := l.svcCtx.ContentRpc.CreatePost(l.ctx, &pb.CreatePostReq{
		UserId:     userId,
		Title:      req.Title,
		Cover:      req.Cover,
		Visibility: 90,
		Contents:   contents,
		Topics:     req.Topics,
		Tags:       "",
	})
	if err != nil {
		logx.Errorf("调用 RPC CreatePost 失败: %v", err)
		code, msg := errutil.ParseRpcError(err)
		return &types.CreatePostResp{
			Code: code,
			Msg:  msg,
		}, nil
	}

	return &types.CreatePostResp{
		Code:   errutil.Success,
		Msg:    errutil.GetCodeMessage(errutil.Success),
		PostId: rpcResp.PostId,
	}, nil
}

func (l *CreatePostLogic) getUserIdFromContext() (uint64, error) {
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
