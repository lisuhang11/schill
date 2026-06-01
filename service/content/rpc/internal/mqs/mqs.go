package mqs

import (
	"context"

	"SChill/common/kafka"
	"SChill/service/content/rpc/internal/config"
	"SChill/service/content/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/service"
)

func Consumers(c config.Config, ctx context.Context, svcContext *svc.ServiceContext) []service.Service {
	store := kafka.NewIdempotencyStore(svcContext.Redis)
	group := c.KqConsumerConf.Group

	commentCreated, err := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: c.KqConsumerConf.Brokers,
		Group:   group,
		Topics:  []string{c.KqConsumerConf.TopicCommentCreated},
	}, NewCommentCreatedHandler(svcContext), store)
	if err != nil {
		panic("comment-created consumer init failed: " + err.Error())
	}

	commentDeleted, err := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: c.KqConsumerConf.Brokers,
		Group:   group,
		Topics:  []string{c.KqConsumerConf.TopicCommentDeleted},
	}, NewCommentDeletedHandler(svcContext), store)
	if err != nil {
		panic("comment-deleted consumer init failed: " + err.Error())
	}

	postStar, err := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: c.KqConsumerConf.Brokers,
		Group:   group,
		Topics:  []string{c.KqConsumerConf.TopicPostStar},
	}, NewPostStarHandler(svcContext), store)
	if err != nil {
		panic("post-star consumer init failed: " + err.Error())
	}

	postUnstar, err := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: c.KqConsumerConf.Brokers,
		Group:   group,
		Topics:  []string{c.KqConsumerConf.TopicPostUnstar},
	}, NewPostUnstarHandler(svcContext), store)
	if err != nil {
		panic("post-unstar consumer init failed: " + err.Error())
	}

	postCollect, err := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: c.KqConsumerConf.Brokers,
		Group:   group,
		Topics:  []string{c.KqConsumerConf.TopicPostCollect},
	}, NewPostCollectHandler(svcContext), store)
	if err != nil {
		panic("post-collect consumer init failed: " + err.Error())
	}

	postUncollect, err := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: c.KqConsumerConf.Brokers,
		Group:   group,
		Topics:  []string{c.KqConsumerConf.TopicPostUncollect},
	}, NewPostUncollectHandler(svcContext), store)
	if err != nil {
		panic("post-uncollect consumer init failed: " + err.Error())
	}

	return []service.Service{
		commentCreated,
		commentDeleted,
		postStar,
		postUnstar,
		postCollect,
		postUncollect,
	}
}
