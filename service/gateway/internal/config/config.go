package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf

	Jwt struct {
		AccessSecret string
	}

	UserRpc        zrpc.RpcClientConf
	ContentRpc     zrpc.RpcClientConf
	FeedRpc        zrpc.RpcClientConf
	CommentRpc     zrpc.RpcClientConf
	RelationRpc    zrpc.RpcClientConf
	InteractionRpc zrpc.RpcClientConf
	SearchRpc      zrpc.RpcClientConf
}
