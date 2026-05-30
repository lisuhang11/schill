package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf
	Mysql struct {
		DataSource string
	}
	Kafka struct {
		Brokers []string
		Topic   string
		Group   string
	}
	Elasticsearch struct {
		Hosts    []string
		Username string
		Password string
	}
}
