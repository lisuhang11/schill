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

type DeletePostLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeletePostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeletePostLogic {
	return &DeletePostLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeletePostLogic) DeletePost(req *types.DeletePostReq) (resp *types.DeletePostResp, err error) {
	if err := req.Validate(); err != nil {
		logx.Errorf("删除帖子参数校验失败: %v", err)
		return &types.DeletePostResp{
			Code: errutil.ErrInvalidParams,
			Msg:  errutil.GetCodeMessage(errutil.ErrInvalidParams),
		}, nil
	}

	userId, err := l.getUserIdFromContext()
	if err != nil || userId == 0 {
		return &types.DeletePostResp{
			Code: errutil.ErrUnauthorized,
			Msg:  errutil.GetCodeMessage(errutil.ErrUnauthorized),
		}, nil
	}

	_, err = l.svcCtx.ContentRpc.DeletePost(l.ctx, &pb.DeletePostReq{
		PostId: req.PostId,
		UserId: userId,
	})
	if err != nil {
		logx.Errorf("调用 RPC DeletePost 失败: %v", err)
		code, msg := errutil.ParseRpcError(err)
		return &types.DeletePostResp{
			Code: code,
			Msg:  msg,
		}, nil
	}

	return &types.DeletePostResp{
		Code: errutil.Success,
		Msg:  errutil.GetCodeMessage(errutil.Success),
	}, nil
}

func (l *DeletePostLogic) getUserIdFromContext() (uint64, error) {
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
