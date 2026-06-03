package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	Elasticsearch struct {
		Hosts      []string
		Username   string
		Password   string
		PostIndex  string
		UserIndex  string
		TopicIndex string
	}

	BizRedis struct {
		Host     string
		Port     string
		Password string
		DB       int
	}
}
