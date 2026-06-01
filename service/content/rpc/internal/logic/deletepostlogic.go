package logic

import (
	"context"
	"fmt"
	"time"

	errutil "SChill/common/error"
	"SChill/common/mq"
	"SChill/service/content/rpc/internal/model"
	"SChill/service/content/rpc/internal/svc"
	"SChill/service/content/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type DeletePostLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeletePostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeletePostLogic {
	return &DeletePostLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeletePostLogic) DeletePost(in *pb.DeletePostReq) (*pb.DeletePostResp, error) {
	if in.UserId == 0 || in.PostId == 0 {
		return nil, errutil.RpcBusinessError(errutil.ErrUnauthorized)
	}

	post, err := loadOwnedPost(l.ctx, l.svcCtx.DBWrite, in.PostId, in.UserId)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	if err := l.svcCtx.DBWrite.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(l.ctx).
			Model(&model.Post{}).
			Where("id = ? AND deleted_at IS NULL", in.PostId).
			Updates(map[string]interface{}{
				"deleted_at": &now,
				"updated_at": now,
			}).Error; err != nil {
			return err
		}

		if err := tx.WithContext(l.ctx).
			Model(&model.PostContent{}).
			Where("post_id = ? AND deleted_at IS NULL", in.PostId).
			Update("deleted_at", &now).Error; err != nil {
			return err
		}

		if err := replacePostTopicsTx(l.ctx, tx, in.PostId, nil); err != nil {
			return err
		}

		return nil
	}); err != nil {
		logx.Errorf("delete post failed: postId=%d err=%v", in.PostId, err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	post.DeletedAt = &now
	post.UpdatedAt = now
	invalidatePostCachesByModel(l.ctx, l.svcCtx, post)

	go func(userID, postID uint64) {
		if err := l.svcCtx.KafkaProducer.SendEvent(
			l.svcCtx.Config.KqPusherConf.TopicDeleted,
			fmt.Sprintf("%d", postID),
			"post.deleted",
			"content-rpc",
			"post",
			fmt.Sprintf("%d", postID),
			mq.PostDeletedMessage{UserID: userID, PostID: postID},
		); err != nil {
			logx.Errorf("send post deleted event failed: %v", err)
		}
	}(in.UserId, in.PostId)

	return &pb.DeletePostResp{Success: true}, nil
}
