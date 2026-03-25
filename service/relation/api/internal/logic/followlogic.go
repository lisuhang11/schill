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

type FollowLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewFollowLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FollowLogic {
	return &FollowLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *FollowLogic) Follow(req *types.FollowReq) (resp *types.FollowResp, err error) {
	if err := req.Validate(); err != nil {
		logx.Errorf("关注参数校验失败: %v", err)
		return &types.FollowResp{
			Code: errutil.ErrInvalidParams,
			Msg:  errutil.GetCodeMessage(errutil.ErrInvalidParams),
		}, nil
	}

	userId, err := l.getUserIdFromContext()
	if err != nil {
		return &types.FollowResp{
			Code: errutil.ErrUnauthorized,
			Msg:  errutil.GetCodeMessage(errutil.ErrUnauthorized),
		}, nil
	}

	_, err = l.svcCtx.RelationRpc.Follow(l.ctx, &pb.FollowReq{
		UserId:       userId,
		TargetUserId: req.TargetUserId,
	})
	if err != nil {
		logx.Errorf("调用 RPC Follow 失败: %v", err)
		code, msg := errutil.ParseRpcError(err)
		return &types.FollowResp{
			Code: code,
			Msg:  msg,
		}, nil
	}

	return &types.FollowResp{
		Code: errutil.Success,
		Msg:  errutil.GetCodeMessage(errutil.Success),
	}, nil
}

func (l *FollowLogic) getUserIdFromContext() (uint64, error) {
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
