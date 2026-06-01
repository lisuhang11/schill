package logic

import (
	"context"
	"fmt"
	"strings"

	errutil "SChill/common/error"
	"SChill/common/mq"
	"SChill/service/content/rpc/internal/model"
	"SChill/service/content/rpc/internal/svc"
	"SChill/service/content/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CreatePostLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreatePostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePostLogic {
	return &CreatePostLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreatePostLogic) CreatePost(in *pb.CreatePostReq) (*pb.CreatePostResp, error) {
	if in.UserId == 0 {
		return nil, errutil.RpcBusinessError(errutil.ErrUnauthorized)
	}

	title := strings.TrimSpace(in.Title)
	if title == "" {
		title = deriveTitleFromContentItems(in.Contents)
	}
	if title == "" {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}

	contents := normalizePostContents(in.Contents)
	if len(contents) == 0 {
		return nil, errutil.RpcBusinessError(errutil.ErrInvalidParams)
	}

	post := &model.Post{
		UserID:     in.UserId,
		Title:      title,
		Cover:      strings.TrimSpace(in.Cover),
		Visibility: in.Visibility,
		Tags:       strings.TrimSpace(in.Tags),
	}
	if post.Cover == "" {
		post.Cover = deriveCoverFromContentItems(contents)
	}

	topicNames := normalizeTopicNames(in.Topics)

	if err := l.svcCtx.DBWrite.WithContext(l.ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(post).Error; err != nil {
			return err
		}

		postContents := make([]model.PostContent, 0, len(contents))
		for idx, item := range contents {
			sort := item.Sort
			if sort <= 0 {
				sort = int32(idx + 1)
			}
			postContents = append(postContents, model.PostContent{
				PostID:  post.ID,
				UserID:  in.UserId,
				Content: item.Content,
				Type:    item.Type,
				Sort:    sort,
			})
		}
		if err := tx.Create(&postContents).Error; err != nil {
			return err
		}

		for _, topicName := range topicNames {
			topic := model.Topic{Name: topicName}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&topic).Error; err != nil {
				return err
			}
			if err := tx.Where("name = ? AND deleted_at IS NULL", topicName).First(&topic).Error; err != nil {
				return err
			}
			if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&model.PostTopic{
				PostID:  post.ID,
				TopicID: topic.ID,
			}).Error; err != nil {
				return err
			}
			if err := tx.Model(&model.Topic{}).
				Where("id = ?", topic.ID).
				Update("quote_num", gorm.Expr("quote_num + ?", 1)).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		logx.Errorf("create post failed: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	invalidatePostCachesByModel(l.ctx, l.svcCtx, post)
	go func(userID, postID uint64) {
		if err := l.svcCtx.KafkaProducer.SendEvent(
			l.svcCtx.Config.KqPusherConf.TopicCreated,
			fmt.Sprintf("%d", postID),
			"post.created",
			"content-rpc",
			"post",
			fmt.Sprintf("%d", postID),
			mq.PostCreatedMessage{UserID: userID, PostID: postID},
		); err != nil {
			logx.Errorf("send post created event failed: %v", err)
		}
	}(in.UserId, post.ID)

	return &pb.CreatePostResp{PostId: post.ID}, nil
}
