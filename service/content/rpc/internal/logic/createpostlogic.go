package logic

import (
	"context"
	"encoding/json"
	"strings"

	errutil "SChill/common/error"
	"SChill/service/content/rpc/internal/model"
	"SChill/service/content/rpc/internal/svc"
	"SChill/service/content/rpc/pb"

	"github.com/IBM/sarama"
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
	summaryItems := make([]contentSummaryItem, 0, len(contents))
	for _, item := range contents {
		summaryItems = append(summaryItems, contentSummaryItem{
			Type:    item.Type,
			Content: item.Content,
		})
	}

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
		payload, err := json.Marshal(struct {
			UserID uint64 `json:"user_id"`
			PostID uint64 `json:"post_id"`
		}{
			UserID: userID,
			PostID: postID,
		})
		if err != nil {
			logx.Errorf("marshal post created event failed: %v", err)
			return
		}

		if _, _, err := l.svcCtx.KafkaProducer.SendMessage(&sarama.ProducerMessage{
			Topic: l.svcCtx.Config.KqPusherConf.TopicCreated,
			Value: sarama.ByteEncoder(payload),
		}); err != nil {
			logx.Errorf("send post created event failed: %v", err)
		}
	}(in.UserId, post.ID)
	publishContentChangedEvent(l.ctx, l.svcCtx, post, "created", buildContentSummaryFromItems(summaryItems), topicNames)

	return &pb.CreatePostResp{PostId: post.ID}, nil
}
