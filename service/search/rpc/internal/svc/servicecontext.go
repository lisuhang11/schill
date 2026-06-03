package svc

import (
	"SChill/common/es"
	"SChill/common/redis"
	"SChill/service/search/rpc/internal/config"
)

type ServiceContext struct {
	Config   config.Config
	ESClient *es.Client
	Redis    *redis.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	esClient, err := es.NewClient(es.Config{
		Hosts:    c.Elasticsearch.Hosts,
		Username: c.Elasticsearch.Username,
		Password: c.Elasticsearch.Password,
	})
	if err != nil {
		panic(err)
	}

	redisClient, err := redis.NewClient(redis.Config{
		Host:     c.BizRedis.Host,
		Port:     c.BizRedis.Port,
		Password: c.BizRedis.Password,
		DB:       c.BizRedis.DB,
	})
	if err != nil {
		panic("Redis client init failed: " + err.Error())
	}

	return &ServiceContext{
		Config:   c,
		ESClient: esClient,
		Redis:    redisClient,
	}
}
