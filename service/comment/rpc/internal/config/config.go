package config

import (
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	ContentRpc zrpc.RpcClientConf

	Mysql struct {
		DataSource      string
		MaxOpenConns    int
		MaxIdleConns    int
		ConnMaxLifetime int
		ConnMaxIdleTime int
		AutoMigrate     bool `json:",optional,default=false"`
	}

	KqProducerConf struct {
		Brokers             []string
		TopicCommentCreate  string
		TopicCommentCreated string
		TopicCommentDeleted string
		TopicCommentVote    string
	}

	KqConsumerConf struct {
		Brokers             []string
		TopicCommentCreate  string
		TopicCommentDeleted string
		TopicCommentVote    string
		TopicCommentDLQ     string
		Group               string
	}

	RedisConf struct {
		Host     string
		Port     string
		Password string
		DB       int
	}

	Cache cache.CacheConf `json:",optional"`
}

func (c Config) MysqlPool() struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
} {
	return struct {
		MaxOpenConns    int
		MaxIdleConns    int
		ConnMaxLifetime time.Duration
		ConnMaxIdleTime time.Duration
	}{
		MaxOpenConns:    c.Mysql.MaxOpenConns,
		MaxIdleConns:    c.Mysql.MaxIdleConns,
		ConnMaxLifetime: time.Duration(c.Mysql.ConnMaxLifetime) * time.Second,
		ConnMaxIdleTime: time.Duration(c.Mysql.ConnMaxIdleTime) * time.Second,
	}
}
