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

type BatchCheckFollowStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewBatchCheckFollowStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *BatchCheckFollowStatusLogic {
	return &BatchCheckFollowStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *BatchCheckFollowStatusLogic) BatchCheckFollowStatus(req *types.BatchCheckFollowStatusReq) (resp *types.BatchCheckFollowStatusResp, err error) {
	if err := req.Validate(); err != nil {
		logx.Errorf("批量检查关注状态参数校验失败: %v", err)
		return &types.BatchCheckFollowStatusResp{
			Code: errutil.ErrInvalidParams,
			Msg:  errutil.GetCodeMessage(errutil.ErrInvalidParams),
		}, nil
	}

	userId, err := l.getUserIdFromContext()
	if err != nil || userId == 0 {
		return &types.BatchCheckFollowStatusResp{
			Code: errutil.ErrUnauthorized,
			Msg:  errutil.GetCodeMessage(errutil.ErrUnauthorized),
		}, nil
	}

	rpcResp, err := l.svcCtx.RelationRpc.BatchCheckFollowStatus(l.ctx, &pb.BatchCheckFollowStatusReq{
		UserId:        userId,
		TargetUserIds: req.TargetUserIds,
	})
	if err != nil {
		logx.Errorf("调用 RPC BatchCheckFollowStatus 失败: %v", err)
		code, msg := errutil.ParseRpcError(err)
		return &types.BatchCheckFollowStatusResp{
			Code: code,
			Msg:  msg,
		}, nil
	}

	var status []types.FollowStatusItem
	for _, item := range rpcResp.Status {
		status = append(status, types.FollowStatusItem{
			UserId:   item.UserId,
			IsFollow: item.IsFollow,
		})
	}

	return &types.BatchCheckFollowStatusResp{
		Code:   errutil.Success,
		Msg:    errutil.GetCodeMessage(errutil.Success),
		Status: status,
	}, nil
}

func (l *BatchCheckFollowStatusLogic) getUserIdFromContext() (uint64, error) {
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
