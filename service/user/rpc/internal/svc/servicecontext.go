package svc

import (
	"errors"
	"time"

	commondb "SChill/common/db"
	commonredis "SChill/common/redis"
	"SChill/service/user/rpc/internal/config"
	"SChill/service/user/rpc/internal/model"

	"github.com/minio/minio-go/v7"
	"github.com/zeromicro/go-zero/core/stores/cache"
	gzredis "github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/syncx"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config      config.Config
	DB          *gorm.DB
	MinIO       *minio.Client
	Cache       cache.Cache
	Redis       *gzredis.Redis
	RedisClient *commonredis.Client
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

	if err := commondb.AutoMigrateIfEnabled(db, c.Mysql.AutoMigrate, &model.User{}, &model.UserProfile{}, &model.UserStat{}); err != nil {
		panic("database migrate failed: " + err.Error())
	}

	if len(c.Cache) == 0 {
		panic("cache nodes not configured")
	}

	redisNode := gzredis.MustNewRedis(c.Cache[0].RedisConf)
	cacheNode := cache.New(
		c.Cache,
		syncx.NewSingleFlight(),
		cache.NewStat("user"),
		errors.New("not found"),
		cache.WithExpiry(time.Duration(commonredis.UserExpire)*time.Second),
		cache.WithNotFoundExpiry(time.Duration(commonredis.CacheNullExpire)*time.Second),
	)

	var redisClient *commonredis.Client
	if c.RedisConf.Host != "" {
		redisClient, err = commonredis.NewClient(commonredis.Config{
			Host:     c.RedisConf.Host,
			Port:     c.RedisConf.Port,
			Password: c.RedisConf.Password,
			DB:       c.RedisConf.DB,
		})
		if err != nil {
			panic("common redis client init failed: " + err.Error())
		}
	}

	return &ServiceContext{
		Config:      c,
		DB:          db,
		Cache:       cacheNode,
		Redis:       redisNode,
		RedisClient: redisClient,
	}
}
