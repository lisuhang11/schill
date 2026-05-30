package mqs

import (
	"context"
	"errors"

	"SChill/service/content/rpc/internal/config"
	"SChill/service/content/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

type managedConsumerService struct {
	name   string
	cancel context.CancelFunc
	run    func() error
}

func newManagedConsumerService(name string, parent context.Context, run func(ctx context.Context) error) service.Service {
	ctx, cancel := context.WithCancel(parent)
	return &managedConsumerService{
		name:   name,
		cancel: cancel,
		run: func() error {
			return run(ctx)
		},
	}
}

func (s *managedConsumerService) Start() {
	go func() {
		if err := s.run(); err != nil && !errors.Is(err, context.Canceled) {
			logx.Errorf("%s exited with error: %v", s.name, err)
		}
	}()
}

func (s *managedConsumerService) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
}

func Consumers(c config.Config, ctx context.Context, svcContext *svc.ServiceContext) []service.Service {
	return []service.Service{
		newManagedConsumerService("content-comment-created-consumer", ctx, func(runCtx context.Context) error {
			return NewCommentCreatedConsumer(runCtx, svcContext).StartConsume(c.KqConsumerConf.Brokers, c.KqConsumerConf.TopicCommentCreated, c.KqConsumerConf.Group)
		}),
		newManagedConsumerService("content-comment-deleted-consumer", ctx, func(runCtx context.Context) error {
			return NewCommentDeletedConsumer(runCtx, svcContext).StartConsume(c.KqConsumerConf.Brokers, c.KqConsumerConf.TopicCommentDeleted, c.KqConsumerConf.Group)
		}),
		newManagedConsumerService("content-post-star-consumer", ctx, func(runCtx context.Context) error {
			return NewPostStarConsumer(runCtx, svcContext).StartConsume(c.KqConsumerConf.Brokers, c.KqConsumerConf.TopicPostStar, c.KqConsumerConf.Group)
		}),
		newManagedConsumerService("content-post-unstar-consumer", ctx, func(runCtx context.Context) error {
			return NewPostUnstarConsumer(runCtx, svcContext).StartConsume(c.KqConsumerConf.Brokers, c.KqConsumerConf.TopicPostUnstar, c.KqConsumerConf.Group)
		}),
		newManagedConsumerService("content-post-collect-consumer", ctx, func(runCtx context.Context) error {
			return NewPostCollectConsumer(runCtx, svcContext).StartConsume(c.KqConsumerConf.Brokers, c.KqConsumerConf.TopicPostCollect, c.KqConsumerConf.Group)
		}),
		newManagedConsumerService("content-post-uncollect-consumer", ctx, func(runCtx context.Context) error {
			return NewPostUncollectConsumer(runCtx, svcContext).StartConsume(c.KqConsumerConf.Brokers, c.KqConsumerConf.TopicPostUncollect, c.KqConsumerConf.Group)
		}),
	}
}
