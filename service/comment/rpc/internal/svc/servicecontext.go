package svc

import (
	commondb "SChill/common/db"
	"SChill/common/mq"
	"SChill/common/redis"
	"SChill/service/comment/rpc/internal/config"
	"SChill/service/comment/rpc/internal/consumers"
	"SChill/service/comment/rpc/internal/model"
	"SChill/service/content/rpc/contentcenter"
	"errors"

	"github.com/zeromicro/go-zero/core/stores/cache"
	gzredis "github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/syncx"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config          config.Config
	DB              *gorm.DB
	Redis           *redis.Client
	Cache           cache.Cache
	KafkaProducer   *mq.Producer
	CommentConsumer *consumers.CommentConsumer
	ContentRpc      contentcenter.ContentCenter
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

	if err := commondb.AutoMigrateIfEnabled(db, c.Mysql.AutoMigrate, &model.Comment{}, &model.CommentContent{}, &model.CommentVote{}, &model.CommentStat{}); err != nil {
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

	kafkaProducer, err := mq.NewProducer(c.KqProducerConf.Brokers)
	if err != nil {
		panic("kafka producer init failed: " + err.Error())
	}

	commentConsumer, err := consumers.NewCommentConsumer(c, db, redisClient, kafkaProducer)
	if err != nil {
		panic("kafka consumer init failed: " + err.Error())
	}

	return &ServiceContext{
		Config:          c,
		DB:              db,
		Redis:           redisClient,
		Cache:           cache.New(cacheConf, syncx.NewSingleFlight(), cache.NewStat("comment"), errors.New("not found")),
		KafkaProducer:   kafkaProducer,
		CommentConsumer: commentConsumer,
		ContentRpc:      contentcenter.NewContentCenter(zrpc.MustNewClient(c.ContentRpc)),
	}
}
