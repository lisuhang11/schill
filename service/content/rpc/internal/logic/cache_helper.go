package logic

import (
	"context"
	"fmt"
	"strconv"

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

func buildPostDetailCacheKey(ctx context.Context, svcCtx *svc.ServiceContext, postID uint64) string {
	return fmt.Sprintf("%s%s:v%s", redis.PostInfoKey, strconv.FormatUint(postID, 10), getPostCacheVersion(ctx, svcCtx, postID))
}

func buildPostContentCacheKey(ctx context.Context, svcCtx *svc.ServiceContext, postID uint64) string {
	return fmt.Sprintf("%s%s:v%s", redis.PostContentKey, strconv.FormatUint(postID, 10), getPostCacheVersion(ctx, svcCtx, postID))
}

func buildPostListCacheKey(ctx context.Context, svcCtx *svc.ServiceContext, userID uint64, feedType string, page, pageSize int64) string {
	return fmt.Sprintf("%suser:%d:feed:%s:page:%d:size:%d:v%s", redis.PostListKey, userID, feedType, page, pageSize, getPostListVersion(ctx, svcCtx, userID))
}

func invalidatePostCaches(ctx context.Context, svcCtx *svc.ServiceContext, postID uint64) {
	_, _ = svcCtx.Redis.Incr(ctx, fmt.Sprintf("%s%d", redis.PostCacheVersionKey, postID))
}

func invalidatePostListCache(ctx context.Context, svcCtx *svc.ServiceContext, userID uint64) {
	_, _ = svcCtx.Redis.Incr(ctx, postListVersionKey(userID))
}

func invalidatePostCachesByModel(ctx context.Context, svcCtx *svc.ServiceContext, post *model.Post) {
	if post == nil {
		return
	}
	invalidatePostCaches(ctx, svcCtx, post.ID)
	invalidatePostListCache(ctx, svcCtx, post.UserID)
}
