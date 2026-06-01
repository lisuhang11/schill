package svc

import (
	"errors"
	"time"

	commondb "SChill/common/db"
	"SChill/common/kafka"
	commonredis "SChill/common/redis"
	"SChill/service/relation/rpc/internal/config"
	"SChill/service/relation/rpc/internal/model"
	"SChill/service/user/rpc/usercenter"

	"github.com/zeromicro/go-zero/core/stores/cache"
	gzredis "github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/syncx"
	"github.com/zeromicro/go-zero/zrpc"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config        config.Config
	DB            *gorm.DB
	KafkaProducer *kafka.Producer
	UserRpc       usercenter.UserCenter
	Cache         cache.Cache
	Redis         *gzredis.Redis
	SingleFlight  syncx.SingleFlight
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(mysql.Open(c.Mysql.DataSource), &gorm.Config{})
	if err != nil {
		panic("database init failed: " + err.Error())
	}

	if err := commondb.AutoMigrateIfEnabled(db, c.Mysql.AutoMigrate, &model.Following{}); err != nil {
		panic("database migrate failed: " + err.Error())
	}

	producer, err := kafka.NewSyncProducer(c.KqPusherConf.Brokers)
	if err != nil {
		panic("kafka producer init failed: " + err.Error())
	}

	if len(c.Cache) == 0 {
		panic("cache nodes not configured")
	}

	barrier := syncx.NewSingleFlight()
	redisNode := gzredis.MustNewRedis(c.Cache[0].RedisConf)
	cacheNode := cache.New(
		c.Cache,
		barrier,
		cache.NewStat("relation"),
		errors.New("not found"),
		cache.WithExpiry(time.Duration(commonredis.RelationCacheExpire)*time.Second),
		cache.WithNotFoundExpiry(time.Duration(commonredis.RelationEmptyExpire)*time.Second),
	)

	return &ServiceContext{
		Config:        c,
		DB:            db,
		KafkaProducer: producer,
		UserRpc:       usercenter.NewUserCenter(zrpc.MustNewClient(c.UserRpc)),
		Cache:         cacheNode,
		Redis:         redisNode,
		SingleFlight:  barrier,
	}
}
