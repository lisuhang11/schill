package svc

import (
	"errors"

	"SChill/common/redis"
	"SChill/service/content/rpc/contentcenter"
	"SChill/service/feed/rpc/internal/config"
	"SChill/service/interaction/rpc/interactioncenter"
	"SChill/service/relation/rpc/relationcenter"
	"SChill/service/user/rpc/usercenter"

	"github.com/zeromicro/go-zero/core/stores/cache"
	gzredis "github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/syncx"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config         config.Config
	ContentRpc     contentcenter.ContentCenter
	UserRpc        usercenter.UserCenter
	RelationRpc    relationcenter.RelationCenter
	InteractionRpc interactioncenter.InteractionCenter
	Redis          *redis.Client
	Cache          cache.Cache
}

func NewServiceContext(c config.Config) *ServiceContext {
	// Initialize Redis client if config is provided
	var redisClient *redis.Client
	var cacheStore cache.Cache

	if c.RedisConf.Host != "" {
		client, err := redis.NewClient(redis.Config{
			Host:     c.RedisConf.Host,
			Port:     c.RedisConf.Port,
			Password: c.RedisConf.Password,
			DB:       c.RedisConf.DB,
		})
		if err != nil {
			panic("feed-rpc redis init failed: " + err.Error())
		}
		redisClient = client

		cacheConf := c.Cache
		if len(cacheConf) == 0 {
			cacheConf = cache.CacheConf{
				{
					RedisConf: gzredis.RedisConf{
						Host: c.RedisConf.Host + ":" + c.RedisConf.Port,
						Pass: c.RedisConf.Password,
					},
				},
			}
		}
		cacheStore = cache.New(cacheConf, syncx.NewSingleFlight(), cache.NewStat("feed"), errors.New("not found"))
	}

	return &ServiceContext{
		Config:         c,
		ContentRpc:     contentcenter.NewContentCenter(zrpc.MustNewClient(c.ContentRpc)),
		UserRpc:        usercenter.NewUserCenter(zrpc.MustNewClient(c.UserRpc)),
		RelationRpc:    relationcenter.NewRelationCenter(zrpc.MustNewClient(c.RelationRpc)),
		InteractionRpc: interactioncenter.NewInteractionCenter(zrpc.MustNewClient(c.InteractionRpc)),
		Redis:          redisClient,
		Cache:          cacheStore,
	}
}
