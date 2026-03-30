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

type PostDeletedMessage struct {
	UserID uint64 `json:"user_id"`
	PostID uint64 `json:"post_id"`
}

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
	if in.PostId == 0 {
		return nil, errutil.RpcBusinessError(errutil.ErrPostNotExist)
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
		var postTopics []model.PostTopic
		err := tx.WithContext(l.ctx).Where("post_id = ?", in.PostId).Find(&postTopics).Error
		if err != nil {
			return err
		}

		for _, pt := range postTopics {
			err = tx.WithContext(l.ctx).Model(&model.Topic{}).Where("id = ?", pt.TopicID).Update("quote_num", gorm.Expr("quote_num - ?", 1)).Error
			if err != nil {
				return err
			}
		}

		err = tx.WithContext(l.ctx).Where("post_id = ?", in.PostId).Delete(&model.PostTopic{}).Error
		if err != nil {
			return err
		}

		err = tx.WithContext(l.ctx).Where("post_id = ?", in.PostId).Delete(&model.PostContent{}).Error
		if err != nil {
			return err
		}

		err = tx.WithContext(l.ctx).Delete(&post).Error
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		logx.Errorf("删除帖子失败: %v", err)
		return nil, errutil.RpcBusinessError(errutil.ErrInternalError)
	}

	msg := PostDeletedMessage{
		UserID: post.UserID,
		PostID: post.ID,
	}
	msgBytes, err := json.Marshal(msg)
	if err != nil {
		logx.Errorf("序列化消息失败: %v", err)
	} else {
		_, _, err := l.svcCtx.KafkaProducer.SendMessage(&sarama.ProducerMessage{
			Topic: l.svcCtx.Config.KqPusherConf.TopicDeleted,
			Value: sarama.StringEncoder(string(msgBytes)),
		})
		if err != nil {
			logx.Errorf("发送Kafka消息失败: %v", err)
		}
	}

	return &pb.DeletePostResp{
		Success: true,
	}, nil
}
