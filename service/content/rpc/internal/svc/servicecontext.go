package svc

import (
	commondb "SChill/common/db"
	"SChill/common/kafka"
	"SChill/common/redis"
	"SChill/service/content/rpc/internal/config"
	"SChill/service/content/rpc/internal/model"
	"SChill/service/relation/rpc/relationcenter"
	"errors"

	"github.com/zeromicro/go-zero/core/collection"
	"github.com/zeromicro/go-zero/core/stores/cache"
	gzredis "github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/syncx"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config        config.Config
	DB            *gorm.DB
	DBRead        *gorm.DB
	DBWrite       *gorm.DB
	Redis         *redis.Client
	Cache         cache.Cache
	LocalCache    *collection.Cache
	KafkaProducer *kafka.Producer
	RelationRpc   relationcenter.RelationCenter
}

func NewServiceContext(c config.Config) *ServiceContext {
	pool := c.MysqlPool()
	db, _, err := commondb.OpenMySQL(c.Mysql.DataSource, commondb.PoolConfig{
		MaxOpenConns:    pool.MaxOpenConns,
		MaxIdleConns:    pool.MaxIdleConns,
		ConnMaxLifetime: pool.ConnMaxLifetime,
		ConnMaxIdleTime: pool.ConnMaxIdleTime,
	})
	if err != nil {
		panic("database init failed: " + err.Error())
	}

	readDB := db
	if c.MysqlRead.DataSource != "" {
		readPool := c.MysqlReadPool()
		readDB, _, err = commondb.OpenMySQL(c.MysqlRead.DataSource, commondb.PoolConfig{
			MaxOpenConns:    readPool.MaxOpenConns,
			MaxIdleConns:    readPool.MaxIdleConns,
			ConnMaxLifetime: readPool.ConnMaxLifetime,
			ConnMaxIdleTime: readPool.ConnMaxIdleTime,
		})
		if err != nil {
			panic("read database init failed: " + err.Error())
		}
	}

	if err := commondb.AutoMigrateIfEnabled(db, c.Mysql.AutoMigrate, &model.Post{}, &model.PostContent{}, &model.PostTopic{}, &model.Topic{}); err != nil {
		panic("database migrate failed: " + err.Error())
	}

	redisClient, err := redis.NewClient(redis.Config{
		Host:     c.RedisConf.Host,
		Port:     c.RedisConf.Port,
		Password: c.RedisConf.Password,
		DB:       c.RedisConf.DB,
	})
	if err != nil {
		panic("redis init failed: " + err.Error())
	}

	cacheConf := c.Cache
	if len(cacheConf) == 0 {
		cacheConf = cache.CacheConf{
			{
				RedisConf: gzredis.RedisConf{
					Host: c.RedisConf.Host + ":" + c.RedisConf.Port,
					Pass: c.RedisConf.Password,
				},
			},
		}
	}

	producer, err := kafka.NewSyncProducer(c.KqPusherConf.Brokers)
	if err != nil {
		panic("kafka producer init failed: " + err.Error())
	}

	relationClient := relationcenter.NewRelationCenter(zrpc.MustNewClient(c.RelationRpcConf))

	return &ServiceContext{
		Config:        c,
		DB:            db,
		DBRead:        readDB,
		DBWrite:       db,
		Redis:         redisClient,
		Cache:         cache.New(cacheConf, syncx.NewSingleFlight(), cache.NewStat("content"), errors.New("not found")),
		LocalCache:    newLocalCache(c.LocalCache.ExpireSeconds, c.LocalCache.Limit),
		KafkaProducer: producer,
		RelationRpc:   relationClient,
	}
}
