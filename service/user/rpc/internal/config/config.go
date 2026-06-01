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

	Jwt struct {
		AccessSecret  string
		AccessExpire  int64
		RefreshSecret string
		RefreshExpire int64
	}

	KqConsumerConf struct {
		Brokers         []string
		Group           string
		TopicCreated    string
		TopicDeleted    string
		TopicFollowed   string
		TopicUnfollowed string
	}

	RedisConf struct {
		Host     string
		Port     string
		Password string
		DB       int
	}

	Cache cache.CacheConf
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
