package svc

import (
	"time"

	"github.com/zeromicro/go-zero/core/collection"
)

func newLocalCache(expireSeconds, limit int) *collection.Cache {
	if expireSeconds <= 0 {
		expireSeconds = 60
	}
	if limit <= 0 {
		limit = 10000
	}

	cache, err := collection.NewCache(
		time.Duration(expireSeconds)*time.Second,
		collection.WithLimit(limit),
		collection.WithName("content-rpc-local"),
	)
	if err != nil {
		panic("local cache init failed: " + err.Error())
	}

	return cache
}
