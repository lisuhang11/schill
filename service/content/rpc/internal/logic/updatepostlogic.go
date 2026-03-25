package logic

import (
	"context"

	errutil "SChill/common/error"
	"SChill/service/content/rpc/internal/model"
	"SChill/service/content/rpc/internal/svc"
	"SChill/service/content/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type UpdatePostLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdatePostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePostLogic {
	return &UpdatePostLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdatePostLogic) UpdatePost(in *pb.UpdatePostReq) (*pb.UpdatePostResp, error) {
	if in.Title == "" {
		return nil, errutil.RpcBusinessError(errutil.ErrPostTitleEmpty)
	}
	if len(in.Contents) == 0 {
		return nil, errutil.RpcBusinessError(errutil.ErrPostContentEmpty)
	}

	var post model.Post
	err := l.svcCtx.DB.WithContext(l.ctx).Where("id = ?", in.PostId).First(&post).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errutil.RpcBusinessError(errutil.ErrPostNotExist)
		}
		logx.Errorf("查询帖子失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}
	if post.UserID != in.UserId {
		return nil, errutil.RpcBusinessError(errutil.ErrNoPermission)
	}

	err = l.svcCtx.DB.Transaction(func(tx *gorm.DB) error {
		var oldPostTopics []model.PostTopic
		if err := tx.WithContext(l.ctx).Where("post_id = ?", in.PostId).Find(&oldPostTopics).Error; err != nil {
			return err
		}

		for _, pt := range oldPostTopics {
			if err := tx.WithContext(l.ctx).Model(&model.Topic{}).Where("id = ?", pt.TopicID).Update("quote_num", gorm.Expr("quote_num - ?", 1)).Error; err != nil {
				return err
			}
		}

		if err := tx.WithContext(l.ctx).Where("post_id = ?", in.PostId).Delete(&model.PostContent{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(l.ctx).Where("post_id = ?", in.PostId).Delete(&model.PostTopic{}).Error; err != nil {
			return err
		}

		post.Visibility = in.Visibility
		post.Tags = in.Tags
		if err := tx.WithContext(l.ctx).Save(&post).Error; err != nil {
			return err
		}

		titleContent := &model.PostContent{
			PostID:  in.PostId,
			UserID:  in.UserId,
			Content: in.Title,
			Type:    1,
			Sort:    1,
		}
		if err := tx.WithContext(l.ctx).Create(titleContent).Error; err != nil {
			return err
		}

		if in.Cover != "" {
			coverContent := &model.PostContent{
				PostID:  in.PostId,
				UserID:  in.UserId,
				Content: in.Cover,
				Type:    3,
				Sort:    2,
			}
			if err := tx.WithContext(l.ctx).Create(coverContent).Error; err != nil {
				return err
			}
		}

		for idx, item := range in.Contents {
			pc := &model.PostContent{
				PostID:  in.PostId,
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
					PostID:  in.PostId,
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
		logx.Errorf("更新帖子失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	return &pb.UpdatePostResp{
		Success: true,
	}, nil
}
