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

type GetFollowingListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFollowingListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFollowingListLogic {
	return &GetFollowingListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFollowingListLogic) GetFollowingList(req *types.GetFollowingListReq) (resp *types.FollowListResp, err error) {
	if err := req.Validate(); err != nil {
		logx.Errorf("获取关注列表参数校验失败: %v", err)
		return &types.FollowListResp{
			Code: errutil.ErrInvalidParams,
			Msg:  errutil.GetCodeMessage(errutil.ErrInvalidParams),
		}, nil
	}

	var userId uint64
	if req.UserId == 0 {
		var getErr error
		userId, getErr = l.getUserIdFromContext()
		if getErr != nil || userId == 0 {
			return &types.FollowListResp{
				Code: errutil.ErrUnauthorized,
				Msg:  errutil.GetCodeMessage(errutil.ErrUnauthorized),
			}, nil
		}
	} else {
		userId = req.UserId
	}

	rpcResp, err := l.svcCtx.RelationRpc.GetFollowingList(l.ctx, &pb.GetFollowingListReq{
		UserId:   userId,
		Page:     req.Page,
		PageSize: req.PageSize,
	})
	if err != nil {
		logx.Errorf("调用 RPC GetFollowingList 失败: %v", err)
		code, msg := errutil.ParseRpcError(err)
		return &types.FollowListResp{
			Code: code,
			Msg:  msg,
		}, nil
	}

	var list []types.FollowInfo
	for _, item := range rpcResp.List {
		list = append(list, types.FollowInfo{
			UserId:     item.UserId,
			Username:   item.Username,
			Avatar:     item.Avatar,
			FollowTime: item.FollowTime,
		})
	}

	return &types.FollowListResp{
		Code:  errutil.Success,
		Msg:   errutil.GetCodeMessage(errutil.Success),
		Total: rpcResp.Total,
		List:  list,
	}, nil
}

func (l *GetFollowingListLogic) getUserIdFromContext() (uint64, error) {
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
