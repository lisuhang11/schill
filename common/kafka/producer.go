package kafka

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"SChill/common/mq"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
)

// ProducerConfig holds configuration for creating a producer.
type ProducerConfig struct {
	Brokers []string
	// Async: if true, uses AsyncProducer (WaitForLocal, Snappy, batch flush).
	// If false, uses SyncProducer (WaitForAll, 5 retries).
	Async bool `json:",optional"`
}

// Producer wraps a sarama producer (Sync or Async) with helper methods.
type Producer struct {
	syncProducer  sarama.SyncProducer
	asyncProducer sarama.AsyncProducer
	async         bool
	closeOnce     sync.Once
	stopChan      chan struct{}
	wg            sync.WaitGroup
}

// NewProducer creates a new Producer.
// For async mode, a background error logger goroutine is started.
func NewProducer(cfg ProducerConfig) (*Producer, error) {
	if cfg.Async {
		return newAsyncProducer(cfg.Brokers)
	}
	return newSyncProducer(cfg.Brokers)
}

// NewSyncProducer creates a producer with WaitForAll acks and 5 retries.
// Use this for critical messages where delivery guarantees matter more than throughput.
func NewSyncProducer(brokers []string) (*Producer, error) {
	return newSyncProducer(brokers)
}

func newSyncProducer(brokers []string) (*Producer, error) {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_8_0_0
	cfg.Producer.RequiredAcks = sarama.WaitForAll
	cfg.Producer.Retry.Max = 5
	cfg.Producer.Return.Successes = true

	prod, err := sarama.NewSyncProducer(brokers, cfg)
	if err != nil {
		return nil, err
	}
	return &Producer{syncProducer: prod, async: false}, nil
}

// NewAsyncProducer creates a producer with WaitForLocal acks and Snappy compression.
// Use this for high-throughput, non-critical messages.
func NewAsyncProducer(brokers []string) (*Producer, error) {
	return newAsyncProducer(brokers)
}

func newAsyncProducer(brokers []string) (*Producer, error) {
	cfg := sarama.NewConfig()
	cfg.Version = sarama.V2_8_0_0
	cfg.Net.MaxOpenRequests = 20
	cfg.Producer.RequiredAcks = sarama.WaitForLocal
	cfg.Producer.Retry.Max = 3
	cfg.Producer.Return.Successes = false
	cfg.Producer.Return.Errors = true
	cfg.Producer.Flush.Frequency = 10 * time.Millisecond
	cfg.Producer.Flush.Messages = 100
	cfg.Producer.MaxMessageBytes = 1048576
	cfg.Producer.Compression = sarama.CompressionSnappy
	cfg.Producer.Idempotent = false

	prod, err := sarama.NewAsyncProducer(brokers, cfg)
	if err != nil {
		return nil, err
	}

	p := &Producer{
		asyncProducer: prod,
		async:         true,
		stopChan:      make(chan struct{}),
	}
	p.wg.Add(1)
	go p.handleAsyncErrors()
	return p, nil
}

func (p *Producer) handleAsyncErrors() {
	defer p.wg.Done()
	for err := range p.asyncProducer.Errors() {
		logx.Errorf("kafka async producer error: topic=%s err=%v", err.Msg.Topic, err.Err)
	}
}

// Send sends a raw JSON message to a topic.
func (p *Producer) Send(topic string, value interface{}) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	return p.sendRaw(topic, "", payload)
}

// SendWithKey sends a raw JSON message to a topic with a partition key.
func (p *Producer) SendWithKey(topic, key string, value interface{}) error {
	payload, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshal message: %w", err)
	}
	return p.sendRaw(topic, key, payload)
}

// SendEvent sends an EventEnvelope-wrapped event to a topic.
// This maintains backward compatibility with existing envelope consumers.
func (p *Producer) SendEvent(topic, key, eventType, producerName, aggregateType, aggregateID string, payload interface{}) error {
	data, _, err := mq.MarshalEnvelope(eventType, producerName, aggregateType, aggregateID, payload)
	if err != nil {
		return fmt.Errorf("marshal envelope: %w", err)
	}
	return p.sendRaw(topic, key, data)
}

// SendRaw sends a pre-built sarama.ProducerMessage directly.
func (p *Producer) SendRaw(msg *sarama.ProducerMessage) error {
	if p.async {
		select {
		case p.asyncProducer.Input() <- msg:
			return nil
		case <-p.stopChan:
			return nil
		}
	}
	_, _, err := p.syncProducer.SendMessage(msg)
	return err
}

func (p *Producer) sendRaw(topic, key string, value []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(value),
	}
	if key != "" {
		msg.Key = sarama.StringEncoder(key)
	}
	return p.SendRaw(msg)
}

// Close closes the producer.
func (p *Producer) Close() error {
	var err error
	p.closeOnce.Do(func() {
		if p.async {
			close(p.stopChan)
			err = p.asyncProducer.Close()
			p.wg.Wait()
		} else {
			err = p.syncProducer.Close()
		}
	})
	return err
}
