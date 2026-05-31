package logic

import (
	"context"
	"fmt"
	"strconv"

	"SChill/common/redis"
	"SChill/service/user/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

func buildUserInfoCacheKey(userID uint64) string {
	return fmt.Sprintf("%s%s", redis.UserInfoKey, strconv.FormatUint(userID, 10))
}

func buildUserProfileCacheKey(userID uint64) string {
	return fmt.Sprintf("%s%s", redis.UserProfileKey, strconv.FormatUint(userID, 10))
}

func buildUserStatCacheKey(userID uint64) string {
	return fmt.Sprintf("%s%s", redis.UserStatKey, strconv.FormatUint(userID, 10))
}

func buildUserBasicInfoCacheKey(userID uint64) string {
	return fmt.Sprintf("%s%s:basic", redis.UserInfoKey, strconv.FormatUint(userID, 10))
}

func invalidateUserCaches(ctx context.Context, svcCtx *svc.ServiceContext, userID uint64) {
	keysToDelete := []string{
		buildUserInfoCacheKey(userID),
		buildUserProfileCacheKey(userID),
		buildUserStatCacheKey(userID),
		buildUserBasicInfoCacheKey(userID),
	}

	for _, key := range keysToDelete {
		if err := svcCtx.Cache.DelCtx(ctx, key); err != nil {
			logx.Errorf("Failed to delete cache key: key=%s, err=%v", key, err)
		}
	}
}

func InvalidateUserCaches(ctx context.Context, svcCtx *svc.ServiceContext, userID uint64) {
	invalidateUserCaches(ctx, svcCtx, userID)
}

// tokenVersionKey returns the Redis key for a user's token version.
// Incrementing this value invalidates all existing refresh tokens for the user.
func tokenVersionKey(userID uint64) string {
	return fmt.Sprintf("%stoken_version:%d", redis.UserInfoKey, userID)
}

// getUserTokenVersion retrieves the current token version for a user from Redis.
// Returns 0 if no version is set (treat as no version check).
func getUserTokenVersion(ctx context.Context, svcCtx *svc.ServiceContext, userID uint64) (int64, error) {
	key := tokenVersionKey(userID)
	val, err := svcCtx.Redis.GetCtx(ctx, key)
	if err != nil {
		if err.Error() == "redis: nil" {
			return 0, nil
		}
		return 0, err
	}
	version, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, nil
	}
	return version, nil
}

// revokeUserTokens increments the token version for a user, invalidating all existing tokens.
// Call this on password change, account freeze, etc.
func revokeUserTokens(ctx context.Context, svcCtx *svc.ServiceContext, userID uint64) error {
	key := tokenVersionKey(userID)
	_, err := svcCtx.Redis.IncrCtx(ctx, key)
	if err != nil {
		logx.Errorf("Failed to increment token version: userId=%d err=%v", userID, err)
		return err
	}
	return nil
}
