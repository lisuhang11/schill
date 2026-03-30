package server

import (
	"context"

	"SChill/service/comment/rpc/internal/logic"
	"SChill/service/comment/rpc/internal/svc"
	"SChill/service/comment/rpc/pb"
)

type CommentCenterServer struct {
	svcCtx *svc.ServiceContext
	pb.UnimplementedCommentCenterServer
}

func NewCommentCenterServer(svcCtx *svc.ServiceContext) *CommentCenterServer {
	return &CommentCenterServer{
		svcCtx: svcCtx,
	}
}

func (s *CommentCenterServer) CreateComment(ctx context.Context, in *pb.CreateCommentReq) (*pb.CreateCommentResp, error) {
	l := logic.NewCreateCommentLogic(ctx, s.svcCtx)
	return l.CreateComment(in)
}

func (s *CommentCenterServer) DeleteComment(ctx context.Context, in *pb.DeleteCommentReq) (*pb.DeleteCommentResp, error) {
	l := logic.NewDeleteCommentLogic(ctx, s.svcCtx)
	return l.DeleteComment(in)
}

func (s *CommentCenterServer) GetCommentList(ctx context.Context, in *pb.GetCommentListReq) (*pb.GetCommentListResp, error) {
	l := logic.NewGetCommentListLogic(ctx, s.svcCtx)
	return l.GetCommentList(in)
}

func (s *CommentCenterServer) GetReplyList(ctx context.Context, in *pb.GetReplyListReq) (*pb.GetReplyListResp, error) {
	l := logic.NewGetReplyListLogic(ctx, s.svcCtx)
	return l.GetReplyList(in)
}

func (s *CommentCenterServer) VoteComment(ctx context.Context, in *pb.VoteCommentReq) (*pb.VoteCommentResp, error) {
	l := logic.NewVoteCommentLogic(ctx, s.svcCtx)
	return l.VoteComment(in)
}
