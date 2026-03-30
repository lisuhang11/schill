package svc

import (
	"SChill/common/mq"
	"SChill/service/comment/rpc/internal/config"
	"SChill/service/comment/rpc/internal/model"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config        config.Config
	DB            *gorm.DB
	KafkaProducer *mq.Producer
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(mysql.Open(c.Mysql.DataSource), &gorm.Config{})
	if err != nil {
		panic("数据库连接失败: " + err.Error())
	}

	err = db.AutoMigrate(&model.Comment{}, &model.CommentContent{}, &model.CommentVote{})
	if err != nil {
		panic("数据库表迁移失败: " + err.Error())
	}

	kafkaProducer, err := mq.NewProducer(c.KqProducerConf.Brokers)
	if err != nil {
		panic("Kafka生产者初始化失败: " + err.Error())
	}

	return &ServiceContext{
		Config:        c,
		DB:            db,
		KafkaProducer: kafkaProducer,
	}
}
