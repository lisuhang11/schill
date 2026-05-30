package mq

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
)

const (
	SchemaVersionV1    = 1
	DefaultEventTTL    = 24 * time.Hour
	IdempotencyKeyPref = "schill:mq:consume:"
)

type EventEnvelope struct {
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`
	AggregateType string          `json:"aggregate_type"`
	AggregateID   string          `json:"aggregate_id"`
	Producer      string          `json:"producer"`
	TraceID       string          `json:"trace_id,omitempty"`
	SchemaVersion int             `json:"schema_version"`
	OccurredAt    int64           `json:"occurred_at"`
	RetryCount    int             `json:"retry_count"`
	Data          json.RawMessage `json:"data"`
}

func BuildEnvelope(eventType, producer, aggregateType, aggregateID string, payload interface{}) (*EventEnvelope, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	now := time.Now().UnixMilli()
	eventID := fmt.Sprintf("%s:%s:%d", eventType, aggregateID, now)
	return &EventEnvelope{
		EventID:       eventID,
		EventType:     eventType,
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		Producer:      producer,
		SchemaVersion: SchemaVersionV1,
		OccurredAt:    now,
		Data:          raw,
	}, nil
}

func MarshalEnvelope(eventType, producer, aggregateType, aggregateID string, payload interface{}) ([]byte, *EventEnvelope, error) {
	envelope, err := BuildEnvelope(eventType, producer, aggregateType, aggregateID, payload)
	if err != nil {
		return nil, nil, err
	}
	data, err := json.Marshal(envelope)
	if err != nil {
		return nil, nil, err
	}
	return data, envelope, nil
}

func DecodeEnvelopePayload(raw []byte, dst interface{}) (*EventEnvelope, error) {
	var envelope EventEnvelope
	if err := json.Unmarshal(raw, &envelope); err == nil && envelope.EventType != "" && len(envelope.Data) > 0 {
		if err := json.Unmarshal(envelope.Data, dst); err != nil {
			return nil, err
		}
		return &envelope, nil
	}

	if err := json.Unmarshal(raw, dst); err != nil {
		return nil, err
	}
	return nil, nil
}

func BuildEnvelopeProducerMessage(topic, key, eventType, producer, aggregateType, aggregateID string, payload interface{}) (*sarama.ProducerMessage, error) {
	data, _, err := MarshalEnvelope(eventType, producer, aggregateType, aggregateID, payload)
	if err != nil {
		return nil, err
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(data),
	}
	if key != "" {
		msg.Key = sarama.StringEncoder(key)
	}
	return msg, nil
}

func BuildIdempotencyKey(group string, envelope *EventEnvelope) string {
	if envelope == nil || envelope.EventID == "" {
		return ""
	}
	return IdempotencyKeyPref + group + ":" + envelope.EventID
}
