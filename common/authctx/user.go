package authctx

import (
	"context"
	"encoding/json"
	"errors"
)

const userIDContextKey = "userId"
const tokenContextKey = "authToken"

var ErrMissingUserID = errors.New("missing user id in context")

func OptionalUserID(ctx context.Context) uint64 {
	if ctx == nil {
		return 0
	}

	switch v := ctx.Value(userIDContextKey).(type) {
	case uint64:
		return v
	case int64:
		if v > 0 {
			return uint64(v)
		}
	case int:
		if v > 0 {
			return uint64(v)
		}
	case float64:
		if v > 0 {
			return uint64(v)
		}
	case json.Number:
		if userID, err := v.Int64(); err == nil && userID > 0 {
			return uint64(userID)
		}
		if userID, err := v.Float64(); err == nil && userID > 0 {
			return uint64(userID)
		}
	}

	return 0
}

func RequireUserID(ctx context.Context) (uint64, error) {
	userID := OptionalUserID(ctx)
	if userID == 0 {
		return 0, ErrMissingUserID
	}
	return userID, nil
}

// TokenFromContext returns the raw JWT token string stored in context (if any).
func TokenFromContext(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	v, _ := ctx.Value(tokenContextKey).(string)
	return v
}
