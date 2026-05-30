package mqs

import (
	"context"
	"errors"

	"SChill/service/user/rpc/internal/config"
	"SChill/service/user/rpc/internal/svc"

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
		newManagedConsumerService("user-post-created-consumer", ctx, func(runCtx context.Context) error {
			return NewPostCreatedConsumer(runCtx, svcContext).StartConsume(c.KqConsumerConf.Brokers, c.KqConsumerConf.TopicCreated, c.KqConsumerConf.Group)
		}),
		newManagedConsumerService("user-post-deleted-consumer", ctx, func(runCtx context.Context) error {
			return NewPostDeletedConsumer(runCtx, svcContext).StartConsume(c.KqConsumerConf.Brokers, c.KqConsumerConf.TopicDeleted, c.KqConsumerConf.Group)
		}),
		newManagedConsumerService("user-followed-consumer", ctx, func(runCtx context.Context) error {
			return NewUserFollowedConsumer(runCtx, svcContext).StartConsume(c.KqConsumerConf.Brokers, c.KqConsumerConf.TopicFollowed, c.KqConsumerConf.Group)
		}),
		newManagedConsumerService("user-unfollowed-consumer", ctx, func(runCtx context.Context) error {
			return NewUserUnfollowedConsumer(runCtx, svcContext).StartConsume(c.KqConsumerConf.Brokers, c.KqConsumerConf.TopicUnfollowed, c.KqConsumerConf.Group)
		}),
	}
}
