package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	Mysql struct {
		DataSource string
	}

	KqPusherConf struct {
		Brokers      []string
		TopicCreated string
		TopicDeleted string
	}

	KqConsumerConf struct {
		Brokers             []string
		Group               string
		TopicCommentCreated string
		TopicCommentDeleted string
	}
}
