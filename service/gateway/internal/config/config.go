package config

import (
	"SChill/common/ratelimit"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf

	Jwt struct {
		AccessSecret string
	}

	// RateLimit enables distributed rate limiting via go-zero PeriodLimit + Redis.
	// When Redis is configured, it is used; otherwise a local in-memory fallback is used.
	RateLimit struct {
		Redis redis.RedisConf
		// Per-path-prefix rate limits.
		Auth       ratelimit.Config
		Write      ratelimit.Config // create/update/delete/vote/follow/star/collect/share
		Read       ratelimit.Config // get/list/feed/search
		Search     ratelimit.Config // search endpoints (stricter)
	}

	UserRpc        zrpc.RpcClientConf
	ContentRpc     zrpc.RpcClientConf
	FeedRpc        zrpc.RpcClientConf
	CommentRpc     zrpc.RpcClientConf
	RelationRpc    zrpc.RpcClientConf
	InteractionRpc zrpc.RpcClientConf
	SearchRpc      zrpc.RpcClientConf
}
