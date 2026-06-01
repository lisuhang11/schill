package kafka

import (
	"context"
	"fmt"
	"time"

	"SChill/common/mq"
	"SChill/common/redis"

	"github.com/zeromicro/go-zero/core/logx"
)

// RedisIdempotencyStore implements IdempotencyStore using Redis SETNX.
type RedisIdempotencyStore struct {
	client *redis.Client
	ttl    time.Duration
}

// NewRedisIdempotencyStore creates a new Redis-backed idempotency store.
func NewRedisIdempotencyStore(client *redis.Client, ttl time.Duration) *RedisIdempotencyStore {
	return &RedisIdempotencyStore{client: client, ttl: ttl}
}

// TryMark marks an event as processed. Returns true if this is the first time,
// false if it has already been processed.
func (s *RedisIdempotencyStore) TryMark(ctx context.Context, group, eventID string) (bool, error) {
	key := fmt.Sprintf("%s%s:%s", mq.IdempotencyKeyPref, group, eventID)
	ok, err := s.client.SetNX(ctx, key, "1", s.ttl)
	if err != nil {
		return false, err
	}
	return ok, nil
}

// NewIdempotencyStore creates a RedisIdempotencyStore with the default TTL (24h).
// Returns nil if client is nil (idempotency disabled).
func NewIdempotencyStore(client *redis.Client) *RedisIdempotencyStore {
	if client == nil {
		return nil
	}
	return NewRedisIdempotencyStore(client, mq.DefaultEventTTL)
}

// NoopIdempotencyStore is an idempotency store that never deduplicates.
type NoopIdempotencyStore struct{}

func (NoopIdempotencyStore) TryMark(ctx context.Context, group, eventID string) (bool, error) {
	return true, nil
}

// logConsumerLag logs the consumer lag for monitoring purposes.
func logConsumerLag(group, topic string, partition int32, offset int64, timestamp time.Time) {
	lag := time.Since(timestamp)
	if lag > 30*time.Second {
		logx.Slowf("kafka consumer lag >30s: group=%s topic=%s partition=%d offset=%d lag_ms=%d",
			group, topic, partition, offset, lag.Milliseconds())
	}
}
