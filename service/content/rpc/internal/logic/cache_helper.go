package logic

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"SChill/common/redis"
	"SChill/service/content/rpc/internal/model"
	"SChill/service/content/rpc/internal/svc"
)

func getPostCacheVersion(ctx context.Context, svcCtx *svc.ServiceContext, postID uint64) string {
	key := fmt.Sprintf("%s%d", redis.PostCacheVersionKey, postID)
	version, err := svcCtx.Redis.Get(ctx, key)
	if err != nil || version == "" {
		return "1"
	}
	return version
}

func postListVersionKey(userID uint64) string {
	return fmt.Sprintf("%s%d", redis.PostListVersionKey, userID)
}

func getPostListVersion(ctx context.Context, svcCtx *svc.ServiceContext, userID uint64) string {
	version, err := svcCtx.Redis.Get(ctx, postListVersionKey(userID))
	if err != nil || version == "" {
		return "1"
	}
	return version
}

// buildPostDetailCacheKey is the legacy combined key, kept for backward compatibility
// during migration. New code should use buildPostBaseCacheKey (in getpostdetaillogic.go).
func buildPostDetailCacheKey(ctx context.Context, svcCtx *svc.ServiceContext, postID uint64) string {
	return fmt.Sprintf("%s%s:v%s", redis.PostInfoKey, strconv.FormatUint(postID, 10), getPostCacheVersion(ctx, svcCtx, postID))
}

func buildPostContentCacheKey(ctx context.Context, svcCtx *svc.ServiceContext, postID uint64) string {
	return fmt.Sprintf("%s%s:v%s", redis.PostContentKey, strconv.FormatUint(postID, 10), getPostCacheVersion(ctx, svcCtx, postID))
}

// buildPostListCacheKey builds the versioned list cache key.
// The version is incremented when a user creates/deletes a post, busting all their list caches.
func buildPostListCacheKey(ctx context.Context, svcCtx *svc.ServiceContext, userID uint64, feedType string, page, pageSize int64) string {
	return fmt.Sprintf("%suser:%d:feed:%s:page:%d:size:%d:v%s", redis.PostListKey, userID, feedType, page, pageSize, getPostListVersion(ctx, svcCtx, userID))
}

// invalidatePostCaches bumps the post cache version, causing the legacy PostInfoKey-based cache to miss.
// Also deletes the new PostBaseKey.
func invalidatePostCaches(ctx context.Context, svcCtx *svc.ServiceContext, postID uint64) {
	_, _ = svcCtx.Redis.Incr(ctx, fmt.Sprintf("%s%d", redis.PostCacheVersionKey, postID))
	_ = svcCtx.Redis.Del(ctx, fmt.Sprintf("%s%d", redis.PostBaseKey, postID))
	_ = svcCtx.Redis.Del(ctx, fmt.Sprintf("%s%d", redis.PostStatsKey, postID))
}

// invalidatePostListCache bumps the list version, busting all list cache keys for a user.
// Uses a timestamp to guarantee a new value every time, even on first invalidation.
// (Incr on a non-existent key returns 1, which is the same as the default version,
// so the first invalidation would be a no-op without using a timestamp.)
func invalidatePostListCache(ctx context.Context, svcCtx *svc.ServiceContext, userID uint64) {
	newVersion := strconv.FormatInt(time.Now().UnixNano(), 10)
	_ = svcCtx.Redis.Set(ctx, postListVersionKey(userID), newVersion, 0)
}

func invalidatePostCachesByModel(ctx context.Context, svcCtx *svc.ServiceContext, post *model.Post) {
	if post == nil {
		return
	}
	invalidatePostCaches(ctx, svcCtx, post.ID)
	invalidatePostListCache(ctx, svcCtx, post.UserID)
}
