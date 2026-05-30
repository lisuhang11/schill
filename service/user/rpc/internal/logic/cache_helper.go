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
