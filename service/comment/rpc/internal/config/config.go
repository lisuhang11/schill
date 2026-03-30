package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	Mysql struct {
		DataSource string
	}

	KqProducerConf struct {
		Brokers         []string
		TopicCommentCreated string
		TopicCommentDeleted string
	}
}
