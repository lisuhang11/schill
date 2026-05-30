package svc

import (
	commondb "SChill/common/db"
	"SChill/common/mq"
	"SChill/common/redis"
	"SChill/service/interaction/rpc/internal/config"
	"SChill/service/interaction/rpc/internal/model"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config            config.Config
	DB                *gorm.DB
	KafkaProducer     *mq.Producer
	KafkaClient       sarama.Client
	PostStarDAO       *model.PostStarDAO
	PostCollectionDAO *model.PostCollectionDAO
	PostShareDAO      *model.PostShareDAO
	RedisClient       *redis.Client
	LuaScripts        *redis.LuaScripts
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

	if err := commondb.AutoMigrateIfEnabled(db, c.Mysql.AutoMigrate, &model.PostStar{}, &model.PostCollection{}, &model.PostShare{}); err != nil {
		panic("database migrate failed: " + err.Error())
	}

	kafkaProducer, err := mq.NewProducer(c.KqProducerConf.Brokers)
	if err != nil {
		panic("kafka producer init failed: " + err.Error())
	}

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Consumer.Return.Errors = true
	kafkaConfig.Version = sarama.V2_8_0_0

	kafkaClient, err := sarama.NewClient(c.KafkaConsumerConf.Brokers, kafkaConfig)
	if err != nil {
		logx.Errorf("failed to connect to Kafka: %v", err)
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

	return &ServiceContext{
		Config:            c,
		DB:                db,
		KafkaProducer:     kafkaProducer,
		KafkaClient:       kafkaClient,
		PostStarDAO:       model.NewPostStarDAO(db),
		PostCollectionDAO: model.NewPostCollectionDAO(db),
		PostShareDAO:      model.NewPostShareDAO(db),
		RedisClient:       redisClient,
		LuaScripts:        redis.NewLuaScripts(),
	}
}
