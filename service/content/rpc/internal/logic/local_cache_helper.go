package logic

import (
	"strings"
	"time"

	"SChill/service/content/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/collection"
)

func loadLocalCache[T any](cache *collection.Cache, key string) (T, bool) {
	var zero T
	if cache == nil || strings.TrimSpace(key) == "" {
		return zero, false
	}

	val, ok := cache.Get(key)
	if !ok {
		return zero, false
	}

	typed, ok := val.(T)
	if !ok {
		return zero, false
	}

	return typed, true
}

func storeLocalCache[T any](svcCtx *svc.ServiceContext, key string, value T, ttl time.Duration) {
	if svcCtx == nil || svcCtx.LocalCache == nil || strings.TrimSpace(key) == "" {
		return
	}
	if ttl <= 0 {
		svcCtx.LocalCache.Set(key, value)
		return
	}

	svcCtx.LocalCache.SetWithExpire(key, value, ttl)
}
