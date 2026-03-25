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

type UpdatePostLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUpdatePostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePostLogic {
	return &UpdatePostLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UpdatePostLogic) UpdatePost(req *types.UpdatePostReq) (resp *types.UpdatePostResp, err error) {
	if err := req.Validate(); err != nil {
		logx.Errorf("更新帖子参数校验失败: %v", err)
		return &types.UpdatePostResp{
			Code: errutil.ErrInvalidParams,
			Msg:  errutil.GetCodeMessage(errutil.ErrInvalidParams),
		}, nil
	}

	userId, err := l.getUserIdFromContext()
	if err != nil || userId == 0 {
		return &types.UpdatePostResp{
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

	_, err = l.svcCtx.ContentRpc.UpdatePost(l.ctx, &pb.UpdatePostReq{
		PostId:     req.PostId,
		UserId:     userId,
		Title:      req.Title,
		Cover:      req.Cover,
		Visibility: 90,
		Contents:   contents,
		Topics:     req.Topics,
		Tags:       "",
	})
	if err != nil {
		logx.Errorf("调用 RPC UpdatePost 失败: %v", err)
		code, msg := errutil.ParseRpcError(err)
		return &types.UpdatePostResp{
			Code: code,
			Msg:  msg,
		}, nil
	}

	return &types.UpdatePostResp{
		Code: errutil.Success,
		Msg:  errutil.GetCodeMessage(errutil.Success),
	}, nil
}

func (l *UpdatePostLogic) getUserIdFromContext() (uint64, error) {
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
