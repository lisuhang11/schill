// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"
	"encoding/json"

	errutil "SChill/common/error"
	"SChill/service/relation/api/internal/svc"
	"SChill/service/relation/api/internal/types"
	"SChill/service/relation/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnfollowLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUnfollowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnfollowLogic {
	return &UnfollowLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UnfollowLogic) Unfollow(req *types.UnfollowReq) (resp *types.UnfollowResp, err error) {
	if err := req.Validate(); err != nil {
		logx.Errorf("取消关注参数校验失败: %v", err)
		return &types.UnfollowResp{
			Code: errutil.ErrInvalidParams,
			Msg:  errutil.GetCodeMessage(errutil.ErrInvalidParams),
		}, nil
	}

	userId, err := l.getUserIdFromContext()
	if err != nil {
		return &types.UnfollowResp{
			Code: errutil.ErrUnauthorized,
			Msg:  errutil.GetCodeMessage(errutil.ErrUnauthorized),
		}, nil
	}

	_, err = l.svcCtx.RelationRpc.Unfollow(l.ctx, &pb.UnfollowReq{
		UserId:       userId,
		TargetUserId: req.TargetUserId,
	})
	if err != nil {
		logx.Errorf("调用 RPC Unfollow 失败: %v", err)
		code, msg := errutil.ParseRpcError(err)
		return &types.UnfollowResp{
			Code: code,
			Msg:  msg,
		}, nil
	}

	return &types.UnfollowResp{
		Code: errutil.Success,
		Msg:  errutil.GetCodeMessage(errutil.Success),
	}, nil
}

func (l *UnfollowLogic) getUserIdFromContext() (uint64, error) {
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
