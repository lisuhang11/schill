package logic

import (
	"context"

	"SChill/service/user/rpc/internal/svc"
	"SChill/service/user/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserStatLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserStatLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserStatLogic {
	return &GetUserStatLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserStatLogic) GetUserStat(in *pb.GetUserStatReq) (*pb.GetUserStatResp, error) {
	// todo: add your logic here and delete this line

	return &pb.GetUserStatResp{}, nil
}
