package mq

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	defaultFlushFrequency = 10 * time.Millisecond
	defaultFlushMessages  = 100
	defaultBatchSize      = 100
)

type Producer struct {
	producer    sarama.AsyncProducer
	messageChan chan *sarama.ProducerMessage
	wg          sync.WaitGroup
	closeOnce   sync.Once
	stopChan    chan struct{}
}

func NewProducer(brokers []string) (*Producer, error) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0

	config.Net.MaxOpenRequests = 20
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Retry.Max = 3
	config.Producer.Return.Successes = false
	config.Producer.Return.Errors = true
	config.Producer.Flush.Frequency = defaultFlushFrequency
	config.Producer.Flush.Messages = defaultFlushMessages
	config.Producer.MaxMessageBytes = 1048576
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Idempotent = false

	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	p := &Producer{
		producer:    producer,
		messageChan: make(chan *sarama.ProducerMessage, 1000),
		stopChan:    make(chan struct{}),
	}

	p.wg.Add(1)
	go p.handleErrors()

	return p, nil
}

func (p *Producer) handleErrors() {
	defer p.wg.Done()
	for err := range p.producer.Errors() {
		logx.Errorf("Kafka message send failed: Topic=%s, Err=%v", err.Msg.Topic, err.Err)
	}
}

func (p *Producer) SendMessage(topic string, message interface{}) error {
	var key string
	if event, ok := message.(CommentCreateEvent); ok {
		key = fmt.Sprintf("%d", event.CommentID)
	}

	return p.SendMessageWithKey(topic, key, message)
}

func (p *Producer) SendEvent(topic, key, eventType, producer, aggregateType, aggregateID string, message interface{}) error {
	msg, err := BuildEnvelopeProducerMessage(topic, key, eventType, producer, aggregateType, aggregateID, message)
	if err != nil {
		logx.Errorf("Failed to build event envelope: %v", err)
		return err
	}
	return p.SendRawMessage(msg)
}

func (p *Producer) SendMessageWithKey(topic string, key string, message interface{}) error {
	value, err := json.Marshal(message)
	if err != nil {
		logx.Errorf("Failed to serialize message: %v", err)
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(value),
	}
	if key != "" {
		msg.Key = sarama.StringEncoder(key)
	}

	select {
	case p.producer.Input() <- msg:
		return nil
	case <-p.stopChan:
		return nil
	}
}

func (p *Producer) SendRawMessage(msg *sarama.ProducerMessage) error {
	select {
	case p.producer.Input() <- msg:
		return nil
	case <-p.stopChan:
		return nil
	}
}

func (p *Producer) Close() error {
	var err error
	p.closeOnce.Do(func() {
		close(p.stopChan)
		err = p.producer.Close()
		p.wg.Wait()
	})
	return err
}
