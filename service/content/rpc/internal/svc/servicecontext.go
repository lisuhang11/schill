package svc

import (
	"SChill/service/content/rpc/internal/config"
	"SChill/service/content/rpc/internal/model"
	"github.com/IBM/sarama"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config        config.Config
	DB            *gorm.DB
	KafkaProducer sarama.SyncProducer
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(mysql.Open(c.Mysql.DataSource), &gorm.Config{})
	if err != nil {
		panic("数据库连接失败: " + err.Error())
	}

	err = db.AutoMigrate(&model.Post{}, &model.PostContent{}, &model.PostTopic{})
	if err != nil {
		panic("数据库表迁移失败: " + err.Error())
	}

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(c.KqPusherConf.Brokers, config)
	if err != nil {
		panic("Kafka producer连接失败: " + err.Error())
	}

	return &ServiceContext{
		Config:        c,
		DB:            db,
		KafkaProducer: producer,
	}
}
