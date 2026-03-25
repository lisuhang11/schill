// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	errutil "SChill/common/error"
	"SChill/service/content/api/internal/svc"
	"SChill/service/content/api/internal/types"
	"SChill/service/content/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type IncViewCountLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewIncViewCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IncViewCountLogic {
	return &IncViewCountLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *IncViewCountLogic) IncViewCount(req *types.IncViewCountReq) (resp *types.IncViewCountResp, err error) {
	if err := req.Validate(); err != nil {
		logx.Errorf("增加浏览量参数校验失败: %v", err)
		return &types.IncViewCountResp{
			Code: errutil.ErrInvalidParams,
			Msg:  errutil.GetCodeMessage(errutil.ErrInvalidParams),
		}, nil
	}

	_, err = l.svcCtx.ContentRpc.IncViewCount(l.ctx, &pb.IncViewCountReq{
		PostId: req.PostId,
	})
	if err != nil {
		logx.Errorf("调用 RPC IncViewCount 失败: %v", err)
		code, msg := errutil.ParseRpcError(err)
		return &types.IncViewCountResp{
			Code: code,
			Msg:  msg,
		}, nil
	}

	return &types.IncViewCountResp{
		Code: errutil.Success,
		Msg:  errutil.GetCodeMessage(errutil.Success),
	}, nil
}
