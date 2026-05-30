package logic

import (
	"context"
	"fmt"

	errutil "SChill/common/error"
	"SChill/common/mq"
	"SChill/service/comment/rpc/internal/model"
	"SChill/service/comment/rpc/internal/svc"
	"SChill/service/comment/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DeleteCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCommentLogic {
	return &DeleteCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteCommentLogic) DeleteComment(in *pb.DeleteCommentReq) (*pb.DeleteCommentResp, error) {
	if in.CommentId == 0 {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}

	var comment model.Comment
	err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", in.CommentId).First(&comment).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
		}
		logx.Errorf("query comment failed: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	if comment.UserID != in.UserId {
		return nil, errutil.RpcBusinessError(errutil.ErrNoPermission)
	}

	invalidatePostCommentListCache(l.ctx, l.svcCtx, comment.PostID)
	if comment.ParentID > 0 {
		invalidateReplyCache(l.ctx, l.svcCtx, comment.ParentID)
	}

	go func() {
		msg := mq.CommentDeletedMessage{
			PostID:    comment.PostID,
			CommentID: comment.ID,
			UserID:    comment.UserID,
		}
		if err := l.svcCtx.KafkaProducer.SendEvent(
			l.svcCtx.Config.KqProducerConf.TopicCommentDeleted,
			fmt.Sprintf("%d", comment.ID),
			"comment.deleted",
			"comment-rpc",
			"comment",
			fmt.Sprintf("%d", comment.ID),
			msg,
		); err != nil {
			logx.Errorf("send comment deleted event failed: %v", err)
		}
	}()

	return &pb.DeleteCommentResp{
		Success: true,
	}, nil
}
