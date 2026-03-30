package logic

import (
	"context"
	"encoding/json"

	errutil "SChill/common/error"
	"SChill/service/content/rpc/internal/model"
	"SChill/service/content/rpc/internal/svc"
	"SChill/service/content/rpc/pb"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type PostCreatedMessage struct {
	UserID uint64 `json:"user_id"`
	PostID uint64 `json:"post_id"`
}

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
	// 参数校验
	if in.Title == "" {
		return nil, errutil.RpcBusinessError(errutil.ErrPostTitleEmpty)
	}
	if len(in.Contents) == 0 {
		return nil, errutil.RpcBusinessError(errutil.ErrPostContentEmpty)
	}

	// 使用事务保证数据一致性：帖子主表、内容表、话题关联等操作要么全部成功，要么全部回滚
	err := l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		post := &model.Post{
			UserID:     in.UserId,
			Visibility: in.Visibility,
			Tags:       in.Tags,
		}
		// 1. 创建帖子主记录
		if err := tx.WithContext(l.ctx).Create(post).Error; err != nil {
			return err
		}

		titleContent := &model.PostContent{
			PostID:  post.ID,
			UserID:  in.UserId,
			Content: in.Title,
			Type:    1,
			Sort:    1,
		}
		// 2. 创建标题内容（type=1）
		if err := tx.WithContext(l.ctx).Create(titleContent).Error; err != nil {
			return err
		}

		if in.Cover != "" {
			coverContent := &model.PostContent{
				PostID:  post.ID,
				UserID:  in.UserId,
				Content: in.Cover,
				Type:    3,
				Sort:    2,
			}
			// 3. 如果提供了封面，创建封面内容（type=3）
			if err := tx.WithContext(l.ctx).Create(coverContent).Error; err != nil {
				return err
			}
		}

		// 4. 遍历创建正文内容（type=2），按顺序设置 Sort
		for idx, item := range in.Contents {
			pc := &model.PostContent{
				PostID:  post.ID,
				UserID:  in.UserId,
				Content: item.Content,
				Type:    item.Type,
				Sort:    int32(idx + 10),
			}
			if item.Sort > 0 {
				pc.Sort = item.Sort
			}
			if err := tx.WithContext(l.ctx).Create(pc).Error; err != nil {
				return err
			}
		}
		// 5. 处理话题关联：话题表引用计数增加，并创建帖子-话题关联记录
		if len(in.Topics) > 0 {
			topicMap := make(map[string]bool)
			for _, topicName := range in.Topics {
				if topicName == "" || topicMap[topicName] {
					continue
				}
				topicMap[topicName] = true

				var topic model.Topic
				err := tx.WithContext(l.ctx).Where("name = ?", topicName).First(&topic).Error
				if err != nil {
					if err == gorm.ErrRecordNotFound {
						topic = model.Topic{
							Name:     topicName,
							QuoteNum: 1,
						}
						if err := tx.WithContext(l.ctx).Create(&topic).Error; err != nil {
							return err
						}
					} else {
						return err
					}
				} else {
					if err := tx.WithContext(l.ctx).Model(&topic).Update("quote_num", gorm.Expr("quote_num + ?", 1)).Error; err != nil {
						return err
					}
				}
				postTopic := &model.PostTopic{
					PostID:  post.ID,
					TopicID: topic.ID,
				}
				if err := tx.WithContext(l.ctx).Create(postTopic).Error; err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		logx.Errorf("创建帖子失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	var post model.Post
	err = l.svcCtx.DB.WithContext(l.ctx).Where("user_id = ?", in.UserId).Order("id DESC").First(&post).Error
	if err != nil {
		logx.Errorf("查询创建的帖子失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}
	// 构造 Kafka 消息，通知用户服务更新该用户的发帖数
	msg := PostCreatedMessage{
		UserID: post.UserID,
		PostID: post.ID,
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		logx.Errorf("序列化消息失败: %v", err)
	} else {
		_, _, err := l.svcCtx.KafkaProducer.SendMessage(&sarama.ProducerMessage{
			Topic: l.svcCtx.Config.KqPusherConf.TopicCreated,
			Value: sarama.StringEncoder(string(msgBytes)),
		})
		if err != nil {
			logx.Errorf("发送Kafka消息失败: %v", err)
		}
	}

	return &pb.CreatePostResp{
		PostId: post.ID,
	}, nil
}
