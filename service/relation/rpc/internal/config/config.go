package config

import (
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	UserRpc zrpc.RpcClientConf

	Mysql struct {
		DataSource  string
		AutoMigrate bool `json:",optional,default=false"`
	}

	KqPusherConf struct {
		Brokers           []string
		TopicFollowed     string
		TopicUnfollowed   string
		TopicMutualFollow string
	}

	Cache cache.CacheConf
}
