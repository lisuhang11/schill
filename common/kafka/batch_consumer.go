package kafka

import (
	"context"
	"time"

	"SChill/common/mq"

	"github.com/IBM/sarama"
	"github.com/zeromicro/go-zero/core/logx"
)

// BatchConfig configures batch processing behavior.
type BatchConfig struct {
	MaxSize      int           // Max messages per batch (default: 100)
	FlushTimeout time.Duration // Max time before flushing (default: 500ms)
}

// BatchHandler is the handler interface for batch consumers.
type BatchHandler interface {
	// HandleBatch processes a batch of messages. The handler should NOT call
	// session.MarkMessage — the framework handles that on success.
	// Return nil on success, or an error to indicate failure.
	// On failure, all messages in the batch are sent to DLQ (if configured).
	HandleBatch(ctx context.Context, msgs []BatchMessage) error
}

// BatchMessage wraps a sarama.ConsumerMessage with pre-parsed data.
type BatchMessage struct {
	Raw      *sarama.ConsumerMessage
	Envelope *mq.EventEnvelope
}

// BatchConsumer wraps a Consumer and adds batch processing support.
// It accumulates messages until MaxSize or FlushTimeout, then calls HandleBatch.
type BatchConsumer struct {
	*Consumer
	batchConfig  BatchConfig
	batchHandler BatchHandler
	dlqTopic     string
	producer     *Producer
}

// NewBatchConsumer creates a new batch consumer.
// The base Consumer handles lifecycle and idempotency.
// BatchConsumer adds batch accumulation and flush logic.
func NewBatchConsumer(
	consumerCfg ConsumerConfig,
	batchHandler BatchHandler,
	batchCfg BatchConfig,
	store IdempotencyStore,
	dlqTopic string,
	dlqProducer *Producer,
) (*BatchConsumer, error) {
	if batchCfg.MaxSize <= 0 {
		batchCfg.MaxSize = 100
	}
	if batchCfg.FlushTimeout <= 0 {
		batchCfg.FlushTimeout = 500 * time.Millisecond
	}

	// Create an inner handler that accumulates messages
	inner := &batchConsumerHandler{
		batchHandler: batchHandler,
		batchConfig:  batchCfg,
		dlqTopic:     dlqTopic,
		producer:     dlqProducer,
		group:        consumerCfg.Group,
	}

	consumer, err := NewConsumer(consumerCfg, inner, store)
	if err != nil {
		return nil, err
	}

	return &BatchConsumer{
		Consumer:     consumer,
		batchConfig:  batchCfg,
		batchHandler: batchHandler,
		dlqTopic:     dlqTopic,
		producer:     dlqProducer,
	}, nil
}

// batchConsumerHandler implements ConsumerHandler and does batch accumulation.
type batchConsumerHandler struct {
	batchHandler BatchHandler
	batchConfig  BatchConfig
	dlqTopic     string
	producer     *Producer
	group        string
}

func (h *batchConsumerHandler) Handle(ctx context.Context, msg *sarama.ConsumerMessage, envelope *mq.EventEnvelope) error {
	// This is called for each message by the Consumer framework.
	// We need batch accumulation, so we override ConsumeClaim directly.
	// The Handle method is not used in batch mode — accumulation happens in ConsumeClaim.
	return nil
}

// BatchConsumeClaim is the batch-aware ConsumeClaim implementation.
// It should be used as the consumerGroupHandler's ConsumeClaim.
func BatchConsumeClaim(
	session sarama.ConsumerGroupSession,
	claim sarama.ConsumerGroupClaim,
	batchHandler BatchHandler,
	batchCfg BatchConfig,
	store IdempotencyStore,
	group string,
	dlqTopic string,
	producer *Producer,
) error {
	ctx := context.Background()

	type pendingMsg struct {
		msg      *sarama.ConsumerMessage
		envelope *mq.EventEnvelope
	}

	batch := make([]BatchMessage, 0, batchCfg.MaxSize)
	pending := make([]*pendingMsg, 0, batchCfg.MaxSize)
	lastFlush := time.Now()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		if err := batchHandler.HandleBatch(ctx, batch); err != nil {
			logx.Errorf("batch flush failed: group=%s err=%v", group, err)
			// Send all pending messages to DLQ
			for _, p := range pending {
				sendToDLQ(producer, dlqTopic, p.msg, err.Error())
			}
		} else {
			for _, p := range pending {
				session.MarkMessage(p.msg, "")
			}
		}
		batch = batch[:0]
		pending = pending[:0]
		lastFlush = time.Now()
	}

	for msg := range claim.Messages() {
		// Parse envelope
		var envelope *mq.EventEnvelope
		var raw interface{}
		env, err := mq.DecodeEnvelopePayload(msg.Value, &raw)
		if err != nil {
			logx.Errorf("batch consumer decode failed: topic=%s partition=%d offset=%d err=%v",
				msg.Topic, msg.Partition, msg.Offset, err)
			session.MarkMessage(msg, "")
			continue
		}
		envelope = env

		// Idempotency check
		if store != nil && envelope != nil {
			ok, err := store.TryMark(ctx, group, envelope.EventID)
			if err != nil {
				logx.Errorf("batch consumer idempotency check failed: group=%s eventID=%s err=%v",
					group, envelope.EventID, err)
			} else if !ok {
				session.MarkMessage(msg, "")
				continue
			}
		}

		batch = append(batch, BatchMessage{Raw: msg, Envelope: envelope})
		pending = append(pending, &pendingMsg{msg: msg, envelope: envelope})

		if len(batch) >= batchCfg.MaxSize || time.Since(lastFlush) >= batchCfg.FlushTimeout {
			flush()
		}
	}

	flush()
	return nil
}

func sendToDLQ(producer *Producer, dlqTopic string, msg *sarama.ConsumerMessage, reason string) {
	if producer == nil || dlqTopic == "" {
		return
	}
	if err := producer.SendRaw(&sarama.ProducerMessage{
		Topic: dlqTopic,
		Key:   sarama.ByteEncoder(msg.Key),
		Value: sarama.ByteEncoder(msg.Value),
		Headers: []sarama.RecordHeader{
			{Key: []byte("x-source-topic"), Value: []byte(msg.Topic)},
			{Key: []byte("x-original-offset"), Value: []byte(formatInt(msg.Offset))},
			{Key: []byte("x-error"), Value: []byte(reason)},
		},
	}); err != nil {
		logx.Errorf("send to DLQ failed: dlq=%s err=%v", dlqTopic, err)
	}
}

func formatInt(n int64) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	b := make([]byte, 0, 20)
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	if neg {
		b = append([]byte{'-'}, b...)
	}
	return string(b)
}
