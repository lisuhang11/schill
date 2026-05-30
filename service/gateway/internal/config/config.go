package config

import (
	"time"

	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf

	Jwt struct {
		AccessSecret string
	}

	SearchProxy struct {
		Target                string
		DialTimeout           time.Duration `json:",default=2s"`
		ResponseHeaderTimeout time.Duration `json:",default=10s"`
	}

	UserRpc        zrpc.RpcClientConf
	ContentRpc     zrpc.RpcClientConf
	FeedRpc        zrpc.RpcClientConf
	CommentRpc     zrpc.RpcClientConf
	RelationRpc    zrpc.RpcClientConf
	InteractionRpc zrpc.RpcClientConf
}
