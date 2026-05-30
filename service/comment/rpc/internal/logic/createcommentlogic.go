package logic

import (
	"context"
	"fmt"
	"time"

	errutil "SChill/common/error"
	"SChill/common/mq"
	"SChill/service/comment/rpc/internal/model"
	"SChill/service/comment/rpc/internal/svc"
	"SChill/service/comment/rpc/pb"
	contentpb "SChill/service/content/rpc/pb"

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
	if in.PostId == 0 || in.Content == "" {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}

	postResp, err := l.svcCtx.ContentRpc.BatchGetPostSummary(l.ctx, &contentpb.BatchGetPostSummaryReq{
		PostIds: []uint64{in.PostId},
	})
	if err != nil {
		logx.Errorf("validate post via content rpc failed: postId=%d err=%v", in.PostId, err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}
	if len(postResp.GetPosts()) == 0 {
		return nil, errutil.RpcBusinessError(errutil.ErrPostNotExist)
	}

	level := int32(1)
	if in.ParentId > 0 {
		var parentComment model.Comment
		err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", in.ParentId).First(&parentComment).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
			}
			logx.Errorf("query parent comment failed: %v", err)
			return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
		}
		level = int32(parentComment.Level + 1)
	}

	var comment *model.Comment
	var createdAt time.Time
	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		var replyToUserID *uint64
		if in.ReplyToUserId > 0 {
			replyToUserID = &in.ReplyToUserId
		}

		createdAt = time.Now()
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
			CreatedAt:     createdAt,
			UpdatedAt:     createdAt,
		}
		if err := tx.WithContext(l.ctx).Create(comment).Error; err != nil {
			return err
		}

		commentContent := &model.CommentContent{
			CommentID:   comment.ID,
			Content:     in.Content,
			ContentType: 1,
			CreatedAt:   createdAt,
			UpdatedAt:   createdAt,
		}
		if err := tx.WithContext(l.ctx).Create(commentContent).Error; err != nil {
			return err
		}

		commentStat := &model.CommentStat{
			CommentID:    comment.ID,
			ReplyCount:   0,
			LikeCount:    0,
			DislikeCount: 0,
		}
		if err := tx.WithContext(l.ctx).Create(commentStat).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		logx.Errorf("create comment failed: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	invalidatePostCommentListCache(l.ctx, l.svcCtx, in.PostId)
	if in.ParentId > 0 {
		invalidateReplyCache(l.ctx, l.svcCtx, in.ParentId)
	}

	respReplyToUserID := uint64(0)
	if comment.ReplyToUserID != nil {
		respReplyToUserID = *comment.ReplyToUserID
	}

	go func() {
		eventMsg := mq.CommentCreateEvent{
			CommentID:     comment.ID,
			PostID:        comment.PostID,
			UserID:        comment.UserID,
			ParentID:      comment.ParentID,
			ReplyToUserID: respReplyToUserID,
			Content:       in.Content,
			CreatedAt:     comment.CreatedAt.Unix(),
			Level:         level,
		}
		if err := l.svcCtx.KafkaProducer.SendEvent(
			l.svcCtx.Config.KqProducerConf.TopicCommentCreate,
			fmt.Sprintf("%d", comment.ID),
			"comment.created.cache_sync",
			"comment-rpc",
			"comment",
			fmt.Sprintf("%d", comment.ID),
			eventMsg,
		); err != nil {
			logx.Errorf("send CommentCreateEvent failed: %v", err)
		}

		createdMsg := mq.CommentCreatedMessage{
			PostID:    comment.PostID,
			CommentID: comment.ID,
			UserID:    comment.UserID,
		}
		if err := l.svcCtx.KafkaProducer.SendEvent(
			l.svcCtx.Config.KqProducerConf.TopicCommentCreated,
			fmt.Sprintf("%d", comment.ID),
			"comment.created.stat_sync",
			"comment-rpc",
			"comment",
			fmt.Sprintf("%d", comment.ID),
			createdMsg,
		); err != nil {
			logx.Errorf("send CommentCreatedMessage failed: %v", err)
		}
	}()

	return &pb.CreateCommentResp{
		Comment: &pb.CommentInfo{
			Id:            comment.ID,
			PostId:        in.PostId,
			UserId:        in.UserId,
			ParentId:      in.ParentId,
			ReplyToUserId: respReplyToUserID,
			Content:       in.Content,
			Level:         level,
			ReplyCount:    0,
			LikeCount:     0,
			DislikeCount:  0,
			CreatedAt:     createdAt.Unix(),
		},
	}, nil
}
