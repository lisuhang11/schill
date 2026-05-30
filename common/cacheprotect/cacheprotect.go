package cacheprotect

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	commonredis "SChill/common/redis"

	"github.com/zeromicro/go-zero/core/syncx"
)

const lockValue = "1"

type Entry struct {
	Data      json.RawMessage `json:"data,omitempty"`
	Empty     bool            `json:"empty"`
	ExpiresAt int64           `json:"expires_at"`
}

func (e *Entry) IsFresh(now time.Time) bool {
	if e == nil {
		return false
	}

	return e.ExpiresAt > now.Unix()
}

func (e *Entry) Decode(v any) error {
	if e == nil || len(e.Data) == 0 {
		return errors.New("cache entry data is empty")
	}

	return json.Unmarshal(e.Data, v)
}

type Group struct {
	once   sync.Once
	flight syncx.SingleFlight
}

func (g *Group) Do(key string, fn func() (interface{}, error)) (interface{}, error, bool) {
	g.init()

	val, fresh, err := g.flight.DoEx(key, func() (any, error) {
		return fn()
	})

	return val, err, fresh
}

func (g *Group) init() {
	g.once.Do(func() {
		g.flight = syncx.NewSingleFlight()
	})
}

func LoadEntry(ctx context.Context, client *commonredis.Client, key string) (*Entry, error) {
	var entry Entry
	if err := client.GetJSON(ctx, key, &entry); err != nil {
		return nil, err
	}

	return &entry, nil
}

func StoreValue(ctx context.Context, client *commonredis.Client, key string, value any, logicalTTL, physicalTTL time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return client.SetJSON(ctx, key, Entry{
		Data:      data,
		Empty:     false,
		ExpiresAt: time.Now().Add(logicalTTL).Unix(),
	}, physicalTTL)
}

func StoreEmpty(ctx context.Context, client *commonredis.Client, key string, logicalTTL, physicalTTL time.Duration) error {
	return client.SetJSON(ctx, key, Entry{
		Empty:     true,
		ExpiresAt: time.Now().Add(logicalTTL).Unix(),
	}, physicalTTL)
}

func StoreMarker(ctx context.Context, client *commonredis.Client, key string, logicalTTL, physicalTTL time.Duration) error {
	return client.SetJSON(ctx, key, Entry{
		Empty:     false,
		ExpiresAt: time.Now().Add(logicalTTL).Unix(),
	}, physicalTTL)
}

func TryLock(ctx context.Context, client *commonredis.Client, key string, ttl time.Duration) (bool, error) {
	return client.SetNX(ctx, key, lockValue, ttl)
}

func ReleaseLock(ctx context.Context, client *commonredis.Client, key string) error {
	return client.Del(ctx, key)
}

func WaitFor(ctx context.Context, attempts int, interval time.Duration, fn func() (bool, error)) (bool, error) {
	if attempts <= 0 {
		return false, nil
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for i := 0; i < attempts; i++ {
		ok, err := fn()
		if ok || err != nil {
			return ok, err
		}

		select {
		case <-ctx.Done():
			return false, ctx.Err()
		case <-ticker.C:
		}
	}

	return false, nil
}
