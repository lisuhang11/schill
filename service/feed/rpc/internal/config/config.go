package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	ContentRpc     zrpc.RpcClientConf
	UserRpc        zrpc.RpcClientConf
	RelationRpc    zrpc.RpcClientConf
	InteractionRpc zrpc.RpcClientConf
}
