package server

import (
	"context"

	"SChill/service/search/rpc/internal/logic"
	"SChill/service/search/rpc/internal/svc"
	"SChill/service/search/rpc/pb"
)

type SearchCenterServer struct {
	svcCtx *svc.ServiceContext
	pb.UnimplementedSearchCenterServer
}

func NewSearchCenterServer(svcCtx *svc.ServiceContext) *SearchCenterServer {
	return &SearchCenterServer{
		svcCtx: svcCtx,
	}
}

func (s *SearchCenterServer) SearchPost(ctx context.Context, in *pb.SearchReq) (*pb.SearchPostResp, error) {
	l := logic.NewSearchPostLogic(ctx, s.svcCtx)
	return l.SearchPost(in)
}

func (s *SearchCenterServer) SearchUser(ctx context.Context, in *pb.SearchReq) (*pb.SearchUserResp, error) {
	l := logic.NewSearchUserLogic(ctx, s.svcCtx)
	return l.SearchUser(in)
}

func (s *SearchCenterServer) SearchTopic(ctx context.Context, in *pb.SearchReq) (*pb.SearchTopicResp, error) {
	l := logic.NewSearchTopicLogic(ctx, s.svcCtx)
	return l.SearchTopic(in)
}
