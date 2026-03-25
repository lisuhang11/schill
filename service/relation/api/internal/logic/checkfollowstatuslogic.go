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

type CheckFollowStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCheckFollowStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CheckFollowStatusLogic {
	return &CheckFollowStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CheckFollowStatusLogic) CheckFollowStatus(req *types.CheckFollowStatusReq) (resp *types.FollowStatusResp, err error) {
	if err := req.Validate(); err != nil {
		logx.Errorf("检查关注状态参数校验失败: %v", err)
		return &types.FollowStatusResp{
			Code: errutil.ErrInvalidParams,
			Msg:  errutil.GetCodeMessage(errutil.ErrInvalidParams),
		}, nil
	}

	userId, err := l.getUserIdFromContext()
	if err != nil {
		return &types.FollowStatusResp{
			Code: errutil.ErrUnauthorized,
			Msg:  errutil.GetCodeMessage(errutil.ErrUnauthorized),
		}, nil
	}

	rpcResp, err := l.svcCtx.RelationRpc.CheckFollowStatus(l.ctx, &pb.CheckFollowStatusReq{
		UserId:       userId,
		TargetUserId: req.TargetUserId,
	})
	if err != nil {
		logx.Errorf("调用 RPC CheckFollowStatus 失败: %v", err)
		code, msg := errutil.ParseRpcError(err)
		return &types.FollowStatusResp{
			Code: code,
			Msg:  msg,
		}, nil
	}

	return &types.FollowStatusResp{
		Code:     errutil.Success,
		Msg:      errutil.GetCodeMessage(errutil.Success),
		IsFollow: rpcResp.IsFollow,
	}, nil
}

func (l *CheckFollowStatusLogic) getUserIdFromContext() (uint64, error) {
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
