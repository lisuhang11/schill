// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"SChill/service/user/api/internal/svc"
	"SChill/service/user/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserStatLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserStatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserStatLogic {
	return &GetUserStatLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserStatLogic) GetUserStat(req *types.GetUserStatReq) (resp *types.GetUserStatResp, err error) {
	// todo: add your logic here and delete this line

	return
}
