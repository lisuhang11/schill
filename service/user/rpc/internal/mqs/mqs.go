package mqs

import (
	"context"

	"SChill/service/user/rpc/internal/config"
	"SChill/service/user/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/service"
)

type PostCreatedConsumerService struct {
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	brokers   []string
	topic     string
	group     string
	consumer  *PostCreatedConsumer
}

type PostDeletedConsumerService struct {
	ctx       context.Context
	svcCtx    *svc.ServiceContext
	brokers   []string
	topic     string
	group     string
	consumer  *PostDeletedConsumer
}

func NewPostCreatedConsumerService(ctx context.Context, svcCtx *svc.ServiceContext, brokers []string, topic string, group string) *PostCreatedConsumerService {
	return &PostCreatedConsumerService{
		ctx:       ctx,
		svcCtx:    svcCtx,
		brokers:   brokers,
		topic:     topic,
		group:     group,
		consumer:  NewPostCreatedConsumer(ctx, svcCtx),
	}
}

func (s *PostCreatedConsumerService) Start() {
	go func() {
		if err := s.consumer.StartConsume(s.brokers, s.topic, s.group); err != nil {
			panic("PostCreated Kafka consumer启动失败: " + err.Error())
		}
	}()
}

func (s *PostCreatedConsumerService) Stop() {
}

func NewPostDeletedConsumerService(ctx context.Context, svcCtx *svc.ServiceContext, brokers []string, topic string, group string) *PostDeletedConsumerService {
	return &PostDeletedConsumerService{
		ctx:       ctx,
		svcCtx:    svcCtx,
		brokers:   brokers,
		topic:     topic,
		group:     group,
		consumer:  NewPostDeletedConsumer(ctx, svcCtx),
	}
}

func (s *PostDeletedConsumerService) Start() {
	go func() {
		if err := s.consumer.StartConsume(s.brokers, s.topic, s.group); err != nil {
			panic("PostDeleted Kafka consumer启动失败: " + err.Error())
		}
	}()
}

func (s *PostDeletedConsumerService) Stop() {
}

func Consumers(c config.Config, ctx context.Context, svcContext *svc.ServiceContext) []service.Service {
	return []service.Service{
		NewPostCreatedConsumerService(
			ctx,
			svcContext,
			c.KqConsumerConf.Brokers,
			c.KqConsumerConf.TopicCreated,
			c.KqConsumerConf.Group,
		),
		NewPostDeletedConsumerService(
			ctx,
			svcContext,
			c.KqConsumerConf.Brokers,
			c.KqConsumerConf.TopicDeleted,
			c.KqConsumerConf.Group,
		),
	}
}
