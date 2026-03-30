package mqs

import (
	"context"

	"SChill/service/content/rpc/internal/config"
	"SChill/service/content/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/service"
)

type CommentCreatedConsumerService struct {
	ctx      context.Context
	svcCtx   *svc.ServiceContext
	brokers  []string
	topic    string
	group    string
	consumer *CommentCreatedConsumer
}

type CommentDeletedConsumerService struct {
	ctx      context.Context
	svcCtx   *svc.ServiceContext
	brokers  []string
	topic    string
	group    string
	consumer *CommentDeletedConsumer
}

func NewCommentCreatedConsumerService(ctx context.Context, svcCtx *svc.ServiceContext, brokers []string, topic string, group string) *CommentCreatedConsumerService {
	return &CommentCreatedConsumerService{
		ctx:      ctx,
		svcCtx:   svcCtx,
		brokers:  brokers,
		topic:    topic,
		group:    group,
		consumer: NewCommentCreatedConsumer(ctx, svcCtx),
	}
}

func (s *CommentCreatedConsumerService) Start() {
	go func() {
		if err := s.consumer.StartConsume(s.brokers, s.topic, s.group); err != nil {
			panic("CommentCreated Kafka consumer启动失败: " + err.Error())
		}
	}()
}

func (s *CommentCreatedConsumerService) Stop() {
}

func NewCommentDeletedConsumerService(ctx context.Context, svcCtx *svc.ServiceContext, brokers []string, topic string, group string) *CommentDeletedConsumerService {
	return &CommentDeletedConsumerService{
		ctx:      ctx,
		svcCtx:   svcCtx,
		brokers:  brokers,
		topic:    topic,
		group:    group,
		consumer: NewCommentDeletedConsumer(ctx, svcCtx),
	}
}

func (s *CommentDeletedConsumerService) Start() {
	go func() {
		if err := s.consumer.StartConsume(s.brokers, s.topic, s.group); err != nil {
			panic("CommentDeleted Kafka consumer启动失败: " + err.Error())
		}
	}()
}

func (s *CommentDeletedConsumerService) Stop() {
}

func Consumers(c config.Config, ctx context.Context, svcContext *svc.ServiceContext) []service.Service {
	return []service.Service{
		NewCommentCreatedConsumerService(
			ctx,
			svcContext,
			c.KqConsumerConf.Brokers,
			c.KqConsumerConf.TopicCommentCreated,
			c.KqConsumerConf.Group,
		),
		NewCommentDeletedConsumerService(
			ctx,
			svcContext,
			c.KqConsumerConf.Brokers,
			c.KqConsumerConf.TopicCommentDeleted,
			c.KqConsumerConf.Group,
		),
	}
}
