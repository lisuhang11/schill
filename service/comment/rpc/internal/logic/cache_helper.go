package logic

import (
	"context"
	"fmt"
	"time"

	"SChill/common/cacheutil"
	"SChill/common/redis"
	"SChill/service/comment/rpc/internal/svc"
)

func buildCommentListKey(postID uint64, sortType string) string {
	if sortType == "hot" {
		return fmt.Sprintf("%s%d:hot", redis.PostCommentsKey, postID)
	}
	return fmt.Sprintf("%s%d:list", redis.PostCommentsKey, postID)
}

func buildCommentListMetaKey(postID uint64, sortType string) string {
	return fmt.Sprintf("%s%d:%s", redis.PostCommentsMetaKey, postID, sortType)
}

func buildCommentListLockKey(postID uint64, sortType string) string {
	return fmt.Sprintf("%spost:%d:%s", redis.CommentLockKey, postID, sortType)
}

func buildReplyListKey(commentID uint64) string {
	return fmt.Sprintf("%s%d:list", redis.CommentRepliesKey, commentID)
}

func buildReplyListMetaKey(commentID uint64) string {
	return fmt.Sprintf("%s%d", redis.CommentRepliesMetaKey, commentID)
}

func buildReplyListLockKey(commentID uint64) string {
	return fmt.Sprintf("%sreply:%d", redis.CommentLockKey, commentID)
}

func invalidatePostCommentListCache(ctx context.Context, svcCtx *svc.ServiceContext, postID uint64) {
	_ = svcCtx.Cache.DelCtx(ctx, fmt.Sprintf("%s%d", redis.PostCommentCountKey, postID))
	_ = svcCtx.Redis.Del(ctx,
		fmt.Sprintf("%s%d:list", redis.PostCommentsKey, postID),
		fmt.Sprintf("%s%d:hot", redis.PostCommentsKey, postID),
		buildCommentListMetaKey(postID, "time"),
		buildCommentListMetaKey(postID, "hot"),
	)
}

func invalidateReplyCache(ctx context.Context, svcCtx *svc.ServiceContext, commentID uint64) {
	_ = svcCtx.Cache.DelCtx(ctx, fmt.Sprintf("%s%d", redis.CommentReplyCountKey, commentID))
	_ = svcCtx.Redis.Del(ctx,
		buildReplyListKey(commentID),
		buildReplyListMetaKey(commentID),
	)
}

func cacheReplyCount(ctx context.Context, svcCtx *svc.ServiceContext, commentID uint64, total int64) {
	ttl := cacheutil.JitterDefault(time.Duration(redis.CommentExpire) * time.Second)
	_ = svcCtx.Cache.SetWithExpireCtx(ctx, fmt.Sprintf("%s%d", redis.CommentReplyCountKey, commentID), total, ttl)
}
