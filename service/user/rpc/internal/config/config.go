package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	Mysql struct {
		DataSource string
	}

	KqConsumerConf struct {
		Brokers      []string
		Group        string
		TopicCreated string
		TopicDeleted string
	}
}
