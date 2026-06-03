// Package ratelimit provides HTTP middleware and utilities for rate limiting
// using go-zero's built-in distributed PeriodLimit, backed by Redis.
//
// Two middleware modes are provided:
//   - Global rate limit per service instance (in-memory, no Redis required).
//   - Distributed rate limit via Redis (PeriodLimit), suitable for multi-instance deployments.
//
// The distributed mode is the default and recommended approach for production,
// as it ensures consistent rate limiting across all gateway replicas.
package ratelimit

import (
	"errors"
	"net/http"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

// Config holds rate limit parameters for a group of routes.
type Config struct {
	// PeriodSec defines the window size in seconds (e.g. 1 = per second).
	PeriodSec int
	// Quota is the maximum number of requests allowed within the window.
	Quota int
}

// RouteLimit maps a path prefix to its rate limit config.
type RouteLimit struct {
	Prefix string
	Config Config
}

// NewPeriodLimiter creates a PeriodLimit-based distributed rate limiter.
// It requires a Redis connection for cross-instance consistency.
// If redisStore is nil, returns an error.
func NewPeriodLimiter(redisStore *redis.Redis, prefix string, cfg Config) (*limit.PeriodLimit, error) {
	if redisStore == nil {
		return nil, errors.New("ratelimit: redis store is required for distributed rate limiting")
	}
	return limit.NewPeriodLimit(cfg.PeriodSec, cfg.Quota, redisStore, prefix), nil
}

// --- In-memory token bucket (fallback / local-only) ---

type tokenBucket struct {
	rate       float64 // tokens per second
	burst      int
	tokens     float64
	lastRefill time.Time
	mu         sync.Mutex
}

func newTokenBucket(rate float64, burst int) *tokenBucket {
	return &tokenBucket{
		rate:       rate,
		burst:      burst,
		tokens:     float64(burst),
		lastRefill: time.Now(),
	}
}

func (tb *tokenBucket) allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * tb.rate
	if tb.tokens > float64(tb.burst) {
		tb.tokens = float64(tb.burst)
	}
	tb.lastRefill = now

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}
	return false
}

// LocalLimiter provides in-memory rate limiting per service instance.
// Use only when Redis is unavailable or for local development.
type LocalLimiter struct {
	buckets map[string]*tokenBucket
	mu      sync.Mutex
	cfg     Config
}

// NewLocalLimiter creates a local in-memory rate limiter.
func NewLocalLimiter(cfg Config) *LocalLimiter {
	return &LocalLimiter{
		buckets: make(map[string]*tokenBucket),
		cfg:     cfg,
	}
}

// Allow checks whether the request identified by key is allowed.
func (l *LocalLimiter) Allow(key string) bool {
	l.mu.Lock()
	tb, ok := l.buckets[key]
	if !ok {
		tb = newTokenBucket(float64(l.cfg.Quota)/float64(l.cfg.PeriodSec), l.cfg.Quota)
		l.buckets[key] = tb
	}
	l.mu.Unlock()
	return tb.allow()
}

// --- HTTP Middleware ---

// Middleware returns a go-zero rest.Middleware that enforces distributed rate limiting.
// limiter can be a *limit.PeriodLimit (distributed) or *LocalLimiter (local).
// name is used in log messages and response headers for identification.
func Middleware(limiter interface{ Take(string) (int, error) }, name string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Use client IP + path prefix as the rate limit key.
			// In production you may want X-Forwarded-For or a user ID.
			key := ClientIP(r)
			code, err := limiter.Take(key)
			if err != nil {
				logx.WithContext(r.Context()).Errorw("ratelimit error",
					logx.Field("key", key),
					logx.Field("err", err),
				)
				// On error, allow the request to avoid blocking legitimate traffic.
				next(w, r)
				return
			}
			switch code {
			case limit.Allowed:
				next(w, r)
			default:
				w.Header().Set("X-RateLimit-Name", name)
				w.Header().Set("Retry-After", "1")
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			}
		}
	}
}

// LocalMiddleware returns a go-zero rest.Middleware that enforces in-memory rate limiting.
func LocalMiddleware(limiter *LocalLimiter, name string) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			key := ClientIP(r)
			if !limiter.Allow(key) {
				w.Header().Set("X-RateLimit-Name", name)
				w.Header().Set("Retry-After", "1")
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
				return
			}
			next(w, r)
		}
	}
}

// ClientIP extracts the client IP from the request.
// Prefers X-Forwarded-For header for proxied requests.
func ClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	return r.RemoteAddr
}
