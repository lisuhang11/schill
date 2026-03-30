package logic

import (
	"context"

	"SChill/service/comment/rpc/internal/svc"
	"SChill/service/comment/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetReplyListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetReplyListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetReplyListLogic {
	return &GetReplyListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetReplyListLogic) GetReplyList(in *pb.GetReplyListReq) (*pb.GetReplyListResp, error) {
	// todo: add your logic here and delete this line

	return &pb.GetReplyListResp{}, nil
}
