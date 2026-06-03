package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"SChill/common/authctx"
	"SChill/common/ratelimit"
	"SChill/service/gateway/internal/config"
	"SChill/service/gateway/internal/handler"
	"SChill/service/gateway/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/limit"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/gateway.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithCustomCors(func(header http.Header) {
		header.Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
	}, nil, "http://localhost:3000"))
	defer server.Stop()

	server.Use(authctx.OptionalJWTMiddleware(c.Jwt.AccessSecret))

	// Rate limit middleware — applied globally, classifies by request path.
	server.Use(rateLimitMiddleware(c))

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting gateway at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

type namedLimiter struct {
	name    string
	limiter interface{ Take(string) (int, error) }
}

func rateLimitMiddleware(c config.Config) rest.Middleware {
	redisStore := newRedisOrNil(c.RateLimit.Redis)

	authLim := namedLimiter{"auth", newPeriodLimitOrLocal(redisStore, "ratelimit:auth", c.RateLimit.Auth)}
	writeLim := namedLimiter{"write", newPeriodLimitOrLocal(redisStore, "ratelimit:write", c.RateLimit.Write)}
	readLim := namedLimiter{"read", newPeriodLimitOrLocal(redisStore, "ratelimit:read", c.RateLimit.Read)}
	searchLim := namedLimiter{"search", newPeriodLimitOrLocal(redisStore, "ratelimit:search", c.RateLimit.Search)}

	logx.Infow("rate limiting enabled",
		logx.Field("auth", fmt.Sprintf("%d/%ds", c.RateLimit.Auth.Quota, c.RateLimit.Auth.PeriodSec)),
		logx.Field("write", fmt.Sprintf("%d/%ds", c.RateLimit.Write.Quota, c.RateLimit.Write.PeriodSec)),
		logx.Field("read", fmt.Sprintf("%d/%ds", c.RateLimit.Read.Quota, c.RateLimit.Read.PeriodSec)),
		logx.Field("search", fmt.Sprintf("%d/%ds", c.RateLimit.Search.Quota, c.RateLimit.Search.PeriodSec)),
		logx.Field("redis", redisStore != nil),
	)

	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			nl := classifyRequest(r, authLim, writeLim, readLim, searchLim)
			if nl == nil {
				// No rate limit for this path (e.g. /health).
				next(w, r)
				return
			}

			key := ratelimit.ClientIP(r)
			code, err := nl.limiter.Take(key)
			if err != nil {
				logx.WithContext(r.Context()).Errorw("ratelimit error",
					logx.Field("name", nl.name),
					logx.Field("key", key),
					logx.Field("err", err),
				)
				// On error, allow the request (fail-open).
				next(w, r)
				return
			}
			switch code {
			case limit.Allowed:
				next(w, r)
			default:
				w.Header().Set("X-RateLimit-Name", nl.name)
				w.Header().Set("Retry-After", "1")
				http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
			}
		}
	}
}

// classifyRequest selects the appropriate rate limiter based on request path.
// Returns nil for paths that should not be rate-limited.
func classifyRequest(r *http.Request, auth, write, read, search namedLimiter) *namedLimiter {
	path := r.URL.Path

	// /health is unauthenticated liveness probe — never rate-limit.
	if path == "/health" {
		return nil
	}

	// Auth endpoints — login/register/refresh.
	if strings.HasPrefix(path, "/api/auth/") {
		return &auth
	}

	// Search endpoints — protect Elasticsearch.
	if strings.HasPrefix(path, "/api/search/") {
		return &search
	}

	// Write endpoints (POST/PUT/DELETE on core resources).
	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch:
		return &write
	}

	// Remaining GET requests — browsing.
	return &read
}

// --- Helpers ---

// newRedisOrNil attempts to create a Redis client from config. Returns nil on failure.
func newRedisOrNil(cfg redis.RedisConf) *redis.Redis {
	if cfg.Host == "" {
		logx.Info("rate limit: no Redis configured, using in-memory fallback")
		return nil
	}
	store := redis.MustNewRedis(cfg)
	if store == nil {
		logx.Info("rate limit: failed to connect Redis, using in-memory fallback")
	}
	return store
}

// newPeriodLimitOrLocal returns a distributed PeriodLimit when redisStore is available,
// otherwise returns a local in-memory limiter.
func newPeriodLimitOrLocal(redisStore *redis.Redis, prefix string, cfg ratelimit.Config) interface {
	Take(string) (int, error)
} {
	if redisStore != nil {
		pl, err := ratelimit.NewPeriodLimiter(redisStore, prefix, cfg)
		if err != nil {
			logx.Errorw("failed to create period limiter, falling back to local", logx.Field("err", err))
			return newLocalAdapter(ratelimit.NewLocalLimiter(cfg))
		}
		return pl
	}
	return newLocalAdapter(ratelimit.NewLocalLimiter(cfg))
}

// localAdapter wraps LocalLimiter to satisfy the Take(string)(int,error) interface.
type localAdapter struct {
	l *ratelimit.LocalLimiter
}

func newLocalAdapter(l *ratelimit.LocalLimiter) *localAdapter {
	return &localAdapter{l: l}
}

func (a *localAdapter) Take(key string) (int, error) {
	const (
		allowed  = 0
		hitQuota = -1
	)
	if a.l.Allow(key) {
		return allowed, nil
	}
	return hitQuota, nil
}
