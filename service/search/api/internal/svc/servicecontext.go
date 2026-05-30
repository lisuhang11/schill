package svc

import (
	"SChill/common/es"
	"SChill/common/redis"
	"SChill/service/search/api/internal/config"
)

type ServiceContext struct {
	Config      config.Config
	ESClient    *es.Client
	Redis       *redis.Client
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
		Host:     c.Redis.Host,
		Port:     c.Redis.Port,
		Password: c.Redis.Password,
		DB:       c.Redis.DB,
	})
	if err != nil {
		panic("Redis 客户端初始化失败: " + err.Error())
	}

	return &ServiceContext{
		Config:      c,
		ESClient:    esClient,
		Redis:       redisClient,
	}
}
