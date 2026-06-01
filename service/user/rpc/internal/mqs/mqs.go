package mqs

import (
	"SChill/common/kafka"
	"SChill/service/user/rpc/internal/config"
	"SChill/service/user/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/service"
)

func Consumers(c config.Config, svcContext *svc.ServiceContext) []service.Service {
	store := kafka.NewIdempotencyStore(svcContext.RedisClient)

	createdConsumer, err := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: c.KqConsumerConf.Brokers,
		Group:   c.KqConsumerConf.Group + "-created",
		Topics:  []string{c.KqConsumerConf.TopicCreated},
	}, NewPostCreatedHandler(svcContext), store)
	if err != nil {
		panic("create post-created consumer failed: " + err.Error())
	}

	deletedConsumer, err := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: c.KqConsumerConf.Brokers,
		Group:   c.KqConsumerConf.Group + "-deleted",
		Topics:  []string{c.KqConsumerConf.TopicDeleted},
	}, NewPostDeletedHandler(svcContext), store)
	if err != nil {
		panic("create post-deleted consumer failed: " + err.Error())
	}

	followedConsumer, err := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: c.KqConsumerConf.Brokers,
		Group:   c.KqConsumerConf.Group + "-followed",
		Topics:  []string{c.KqConsumerConf.TopicFollowed},
	}, NewUserFollowedHandler(svcContext), store)
	if err != nil {
		panic("create user-followed consumer failed: " + err.Error())
	}

	unfollowedConsumer, err := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers: c.KqConsumerConf.Brokers,
		Group:   c.KqConsumerConf.Group + "-unfollowed",
		Topics:  []string{c.KqConsumerConf.TopicUnfollowed},
	}, NewUserUnfollowedHandler(svcContext), store)
	if err != nil {
		panic("create user-unfollowed consumer failed: " + err.Error())
	}

	return []service.Service{createdConsumer, deletedConsumer, followedConsumer, unfollowedConsumer}
}
