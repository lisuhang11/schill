package kafka

import (
	"context"
	"errors"
	"sync"
	"time"

	"SChill/common/mq"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/service"
)

// ConsumerConfig holds configuration for a Kafka consumer group.
type ConsumerConfig struct {
	Brokers []string
	Group   string
	Topics  []string
	// OffsetInitial controls where to start consuming: "oldest" (default) or "newest".
	OffsetInitial string `json:",optional"`
}

// ConsumerHandler is the interface that consumer handlers must implement.
// It is called for each message; the handler is responsible for processing
// and returning an error if processing failed.
type ConsumerHandler interface {
	// Handle processes a single message. The envelope is parsed from the message
	// if it was wrapped; otherwise it's nil. Return nil to mark the message as
	// processed, or an error to indicate failure.
	Handle(ctx context.Context, msg *sarama.ConsumerMessage, envelope *mq.EventEnvelope) error
}

// Consumer wraps a sarama.ConsumerGroup and implements service.Service.
type Consumer struct {
	config    ConsumerConfig
	handler   ConsumerHandler
	group     sarama.ConsumerGroup
	ctx       context.Context
	cancel    context.CancelFunc
	wg        sync.WaitGroup

	// Idempotency store (optional). If set, messages with the same EventEnvelope.EventID
	// will be skipped. Uses Redis SETNX under the hood.
	idempotencyStore IdempotencyStore
}

// IdempotencyStore provides deduplication using a distributed store.
type IdempotencyStore interface {
	// TryMark returns true if the event ID was NOT previously seen and has been marked.
	// Returns false if the event has already been processed.
	TryMark(ctx context.Context, group, eventID string) (bool, error)
}

// NewConsumer creates a new Kafka consumer group.
func NewConsumer(cfg ConsumerConfig, handler ConsumerHandler, store IdempotencyStore) (*Consumer, error) {
	ctx, cancel := context.WithCancel(context.Background())

	scfg := sarama.NewConfig()
	scfg.Version = sarama.V2_8_0_0
	scfg.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin

	offsetInitial := sarama.OffsetOldest
	if cfg.OffsetInitial == "newest" {
		offsetInitial = sarama.OffsetNewest
	}
	scfg.Consumer.Offsets.Initial = offsetInitial
	scfg.Consumer.Return.Errors = true

	group, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.Group, scfg)
	if err != nil {
		cancel()
		return nil, err
	}

	return &Consumer{
		config:           cfg,
		handler:          handler,
		group:            group,
		ctx:              ctx,
		cancel:           cancel,
		idempotencyStore: store,
	}, nil
}

// Start starts the consumer in a background goroutine.
func (c *Consumer) Start() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		for {
			if err := c.group.Consume(c.ctx, c.config.Topics, &consumerGroupHandler{
				handler:          c.handler,
				group:            c.config.Group,
				idempotencyStore: c.idempotencyStore,
			}); err != nil {
				if !errors.Is(err, context.Canceled) {
					logx.Errorf("kafka consumer error: group=%s err=%v", c.config.Group, err)
				}
			}
			if c.ctx.Err() != nil {
				return
			}
			// Brief delay before reconnecting
			time.Sleep(time.Second)
		}
	}()
	logx.Infof("kafka consumer started: group=%s topics=%v", c.config.Group, c.config.Topics)
}

// Stop gracefully stops the consumer.
func (c *Consumer) Stop() {
	c.cancel()
	if err := c.group.Close(); err != nil {
		logx.Errorf("kafka consumer close error: group=%s err=%v", c.config.Group, err)
	}
	c.wg.Wait()
}

// consumerGroupHandler implements sarama.ConsumerGroupHandler.
type consumerGroupHandler struct {
	handler          ConsumerHandler
	group            string
	idempotencyStore IdempotencyStore
}

func (h *consumerGroupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *consumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *consumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := context.Background()
	for msg := range claim.Messages() {
		lag := time.Since(msg.Timestamp)
		logx.Infof("kafka consumer received: topic=%s partition=%d offset=%d lag_ms=%d",
			msg.Topic, msg.Partition, msg.Offset, lag.Milliseconds())

		// Parse envelope if present
		var envelope *mq.EventEnvelope
		var raw interface{}
		envelope, err := mq.DecodeEnvelopePayload(msg.Value, &raw)
		if err != nil {
			logx.Errorf("kafka consumer decode failed: topic=%s partition=%d offset=%d err=%v",
				msg.Topic, msg.Partition, msg.Offset, err)
			session.MarkMessage(msg, "")
			continue
		}

		// Idempotency check
		if h.idempotencyStore != nil && envelope != nil {
			ok, err := h.idempotencyStore.TryMark(ctx, h.group, envelope.EventID)
			if err != nil {
				logx.Errorf("kafka consumer idempotency check failed: group=%s eventID=%s err=%v",
					h.group, envelope.EventID, err)
			} else if !ok {
				// Already processed, skip
				session.MarkMessage(msg, "")
				continue
			}
		}

		if err := h.handler.Handle(ctx, msg, envelope); err != nil {
			logx.Errorf("kafka consumer handle failed: topic=%s partition=%d offset=%d err=%v",
				msg.Topic, msg.Partition, msg.Offset, err)
			// Don't mark — let the consumer retry on restart or after rebalance
			continue
		}

		session.MarkMessage(msg, "")
	}
	return nil
}

// Compile-time check: Consumer implements service.Service.
var _ service.Service = (*Consumer)(nil)
