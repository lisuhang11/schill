package config

import (
	"time"

	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf

	Mysql struct {
		DataSource      string
		MaxOpenConns    int
		MaxIdleConns    int
		ConnMaxLifetime int
		ConnMaxIdleTime int
		AutoMigrate     bool `json:",optional,default=false"`
	}

	MysqlRead struct {
		DataSource      string `json:",optional"`
		MaxOpenConns    int    `json:",optional"`
		MaxIdleConns    int    `json:",optional"`
		ConnMaxLifetime int    `json:",optional"`
		ConnMaxIdleTime int    `json:",optional"`
	}

	RelationRpcConf zrpc.RpcClientConf

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
		TopicPostStar       string
		TopicPostUnstar     string
		TopicPostCollect    string
		TopicPostUncollect  string
	}

	RedisConf struct {
		Host     string
		Port     string
		Password string
		DB       int
	}

	Cache cache.CacheConf `json:",optional"`

	LocalCache struct {
		ExpireSeconds int `json:",optional,default=60"`
		Limit         int `json:",optional,default=10000"`
	}
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

func (c Config) MysqlReadPool() struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
} {
	maxOpenConns := c.MysqlRead.MaxOpenConns
	if maxOpenConns <= 0 {
		maxOpenConns = c.Mysql.MaxOpenConns
	}

	maxIdleConns := c.MysqlRead.MaxIdleConns
	if maxIdleConns <= 0 {
		maxIdleConns = c.Mysql.MaxIdleConns
	}

	connMaxLifetime := c.MysqlRead.ConnMaxLifetime
	if connMaxLifetime <= 0 {
		connMaxLifetime = c.Mysql.ConnMaxLifetime
	}

	connMaxIdleTime := c.MysqlRead.ConnMaxIdleTime
	if connMaxIdleTime <= 0 {
		connMaxIdleTime = c.Mysql.ConnMaxIdleTime
	}

	return struct {
		MaxOpenConns    int
		MaxIdleConns    int
		ConnMaxLifetime time.Duration
		ConnMaxIdleTime time.Duration
	}{
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: time.Duration(connMaxLifetime) * time.Second,
		ConnMaxIdleTime: time.Duration(connMaxIdleTime) * time.Second,
	}
}
