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

type CreateCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCommentLogic {
	return &CreateCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateCommentLogic) CreateComment(in *pb.CreateCommentReq) (*pb.CreateCommentResp, error) {
	if in.PostId == 0 {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}
	if in.Content == "" {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}

	var postExists struct {
		ID uint64
	}
	err := l.svcCtx.DB.WithContext(l.ctx).Table("post").Select("id").Where("id = ?", in.PostId).First(&postExists).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errutil.RpcBusinessError(errutil.ErrPostNotExist)
		}
		logx.Errorf("查询帖子失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	level := int32(1)
	if in.ParentId > 0 {
		var parentComment model.Comment
		err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", in.ParentId).First(&parentComment).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
			}
			logx.Errorf("查询父评论失败: %v", err)
			return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
		}
		level = int32(parentComment.Level + 1)
	}

	var comment *model.Comment
	var commentContent *model.CommentContent

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		var replyToUserID *uint64
		if in.ReplyToUserId > 0 {
			replyToUserID = &in.ReplyToUserId
		}

		comment = &model.Comment{
			PostID:        in.PostId,
			UserID:        in.UserId,
			ParentID:      in.ParentId,
			ReplyToUserID: replyToUserID,
			Level:         uint8(level),
			Status:        1,
			ReplyCount:    0,
			LikeCount:     0,
			DislikeCount:  0,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if err := tx.WithContext(l.ctx).Create(comment).Error; err != nil {
			return err
		}

		commentContent = &model.CommentContent{
			CommentID:   comment.ID,
			Content:     in.Content,
			ContentType: 1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if err := tx.WithContext(l.ctx).Create(commentContent).Error; err != nil {
			return err
		}

		if in.ParentId > 0 {
			if err := tx.WithContext(l.ctx).Model(&model.Comment{}).
				Where("id = ?", in.ParentId).
				Update("reply_count", gorm.Expr("reply_count + ?", 1)).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		logx.Errorf("创建评论失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	respReplyToUserID := uint64(0)
	if comment.ReplyToUserID != nil {
		respReplyToUserID = *comment.ReplyToUserID
	}

	go func() {
		msg := mq.CommentCreatedMessage{
			PostID:    comment.PostID,
			CommentID: comment.ID,
			UserID:    comment.UserID,
		}
		if err := l.svcCtx.KafkaProducer.SendMessage(l.svcCtx.Config.KqProducerConf.TopicCommentCreated, msg); err != nil {
			logx.Errorf("发送评论创建消息失败: %v", err)
		}
	}()

	return &pb.CreateCommentResp{
		Comment: &pb.CommentInfo{
			Id:           comment.ID,
			PostId:       comment.PostID,
			UserId:       comment.UserID,
			ParentId:     comment.ParentID,
			ReplyToUserId: respReplyToUserID,
			Content:      commentContent.Content,
			Level:        level,
			ReplyCount:   comment.ReplyCount,
			LikeCount:    comment.LikeCount,
			CreatedAt:    comment.CreatedAt.Unix(),
		},
	}, nil
}
