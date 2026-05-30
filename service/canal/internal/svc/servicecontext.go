package svc

import (
	commondb "SChill/common/db"
	"SChill/common/es"
	"SChill/service/canal/internal/config"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config      config.Config
	ESClient    *es.Client
	KafkaClient sarama.Client
	DB          *gorm.DB
}

func NewServiceContext(c config.Config) *ServiceContext {
	esClient, err := es.NewClient(es.Config{
		Hosts:    c.Elasticsearch.Hosts,
		Username: c.Elasticsearch.Username,
		Password: c.Elasticsearch.Password,
	})
	if err != nil {
		logx.Errorf("Failed to connect to Elasticsearch: %v", err)
	}

	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Consumer.Return.Errors = true
	kafkaConfig.Version = sarama.V2_8_0_0

	kafkaClient, err := sarama.NewClient(c.Kafka.Brokers, kafkaConfig)
	if err != nil {
		logx.Errorf("Failed to connect to Kafka: %v", err)
	}

	db, _, err := commondb.OpenMySQL(c.Mysql.DataSource, commondb.PoolConfig{})
	if err != nil {
		logx.Errorf("Failed to connect to MySQL: %v", err)
	}

	return &ServiceContext{
		Config:      c,
		ESClient:    esClient,
		KafkaClient: kafkaClient,
		DB:          db,
	}
}
