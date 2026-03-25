// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package logic

import (
	"context"

	"SChill/service/comment/api/internal/svc"
	"SChill/service/comment/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type VoteCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVoteCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VoteCommentLogic {
	return &VoteCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VoteCommentLogic) VoteComment(req *types.VoteCommentReq) (resp *types.VoteCommentResp, err error) {
	// todo: add your logic here and delete this line

	return
}
