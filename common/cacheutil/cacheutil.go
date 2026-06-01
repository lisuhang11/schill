// Package cacheutil provides unified caching utilities built on go-zero's cache primitives.
//
// It adds:
//   - TTL jitter to prevent thundering herd on key expiry
//   - Layered cache key generation for base/stats/viewerState
//   - Random TTL helpers
package cacheutil

import (
	"crypto/rand"
	"math/big"
	"time"
)

const (
	// DefaultJitterRatio is the ± ratio applied to TTLs for jitter.
	// e.g. 0.1 means a 10min TTL gets ±1min jitter.
	DefaultJitterRatio = 0.1
)

// Jitter returns a duration randomly adjusted by ±ratio.
//
// Example: Jitter(10*time.Minute, 0.1) → [9min, 11min]
func Jitter(base time.Duration, ratio float64) time.Duration {
	if ratio <= 0 || base <= 0 {
		return base
	}
	delta := time.Duration(float64(base) * ratio)
	if delta <= 0 {
		return base
	}
	n, err := rand.Int(rand.Reader, big.NewInt(int64(delta*2+1)))
	if err != nil {
		return base
	}
	return base - delta + time.Duration(n.Int64())
}

// JitterDefault is like Jitter with DefaultJitterRatio (10%).
func JitterDefault(base time.Duration) time.Duration {
	return Jitter(base, DefaultJitterRatio)
}
