package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	ContentRpc     zrpc.RpcClientConf
	UserRpc        zrpc.RpcClientConf
	RelationRpc    zrpc.RpcClientConf
	InteractionRpc zrpc.RpcClientConf

	Cache    cache.CacheConf `json:",optional"`
	RedisConf struct {
		Host     string
		Port     string
		Password string
		DB       int
	} `json:",optional"`
}
