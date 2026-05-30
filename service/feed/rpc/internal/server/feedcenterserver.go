package server

import (
	"context"

	"SChill/service/feed/rpc/internal/logic"
	"SChill/service/feed/rpc/internal/svc"
	"SChill/service/feed/rpc/pb"
)

type FeedCenterServer struct {
	svcCtx *svc.ServiceContext
	pb.UnimplementedFeedCenterServer
}

func NewFeedCenterServer(svcCtx *svc.ServiceContext) *FeedCenterServer {
	return &FeedCenterServer{
		svcCtx: svcCtx,
	}
}

func (s *FeedCenterServer) GetFeedList(ctx context.Context, in *pb.GetFeedListReq) (*pb.GetFeedListResp, error) {
	l := logic.NewGetFeedListLogic(ctx, s.svcCtx)
	return l.GetFeedList(in)
}
