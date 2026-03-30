package mq

import (
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
)

type Producer struct {
	producer sarama.SyncProducer
}

func NewProducer(brokers []string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Producer{
		producer: producer,
	}, nil
}

func (p *Producer) SendMessage(topic string, message interface{}) error {
	value, err := json.Marshal(message)
	if err != nil {
		logx.Errorf("序列化消息失败: %v", err)
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(value),
	}

	partition, offset, err := p.producer.SendMessage(msg)
	if err != nil {
		logx.Errorf("发送Kafka消息失败: Topic=%s, Err=%v", topic, err)
		return err
	}

	logx.Infof("发送Kafka消息成功: Topic=%s, Partition=%d, Offset=%d", topic, partition, offset)
	return nil
}

func (p *Producer) Close() error {
	return p.producer.Close()
}
