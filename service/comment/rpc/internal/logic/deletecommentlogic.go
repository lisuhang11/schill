package logic

import (
	"context"
	"time"

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
		logx.Errorf("查询评论失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	if comment.UserID != in.UserId {
		return nil, errutil.RpcBusinessError(errutil.ErrNoPermission)
	}

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		comment.DeletedAt = &now

		if err := tx.WithContext(l.ctx).Save(&comment).Error; err != nil {
			return err
		}

		if comment.ParentID > 0 {
			if err := tx.WithContext(l.ctx).Model(&model.Comment{}).
				Where("id = ?", comment.ParentID).
				Update("reply_count", gorm.Expr("reply_count - ?", 1)).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		logx.Errorf("删除评论失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	go func() {
		msg := mq.CommentDeletedMessage{
			PostID:    comment.PostID,
			CommentID: comment.ID,
			UserID:    comment.UserID,
		}
		if err := l.svcCtx.KafkaProducer.SendMessage(l.svcCtx.Config.KqProducerConf.TopicCommentDeleted, msg); err != nil {
			logx.Errorf("发送评论删除消息失败: %v", err)
		}
	}()

	return &pb.DeleteCommentResp{
		Success: true,
	}, nil
}
