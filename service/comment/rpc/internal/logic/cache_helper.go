package logic

import (
	"context"
	"fmt"
	"time"

	"SChill/common/cacheutil"
	"SChill/common/redis"
	"SChill/service/comment/rpc/internal/svc"
)

func buildCommentListKey(postID uint64) string {
	return fmt.Sprintf("%s%d:list", redis.PostCommentsKey, postID)
}

func buildCommentListMetaKey(postID uint64) string {
	return fmt.Sprintf("%s%d", redis.PostCommentsMetaKey, postID)
}

func buildCommentListLockKey(postID uint64) string {
	return fmt.Sprintf("%spost:%d", redis.CommentLockKey, postID)
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
	countKey := fmt.Sprintf("%s%d", redis.PostCommentCountKey, postID)
	_ = svcCtx.Redis.Del(ctx, countKey)
	_ = svcCtx.Redis.Del(ctx,
		buildCommentListKey(postID),
		buildCommentListMetaKey(postID),
	)
}

func invalidateReplyCache(ctx context.Context, svcCtx *svc.ServiceContext, commentID uint64) {
	countKey := fmt.Sprintf("%s%d", redis.CommentReplyCountKey, commentID)
	_ = svcCtx.Redis.Del(ctx, countKey)
	_ = svcCtx.Redis.Del(ctx,
		buildReplyListKey(commentID),
		buildReplyListMetaKey(commentID),
	)
}

func cacheCommentCount(ctx context.Context, svcCtx *svc.ServiceContext, postID uint64, total int64) {
	ttl := cacheutil.JitterDefault(time.Duration(redis.CommentExpire) * time.Second)
	_ = svcCtx.Redis.Set(ctx, fmt.Sprintf("%s%d", redis.PostCommentCountKey, postID), fmt.Sprintf("%d", total), ttl)
}

func cacheReplyCount(ctx context.Context, svcCtx *svc.ServiceContext, commentID uint64, total int64) {
	ttl := cacheutil.JitterDefault(time.Duration(redis.CommentExpire) * time.Second)
	_ = svcCtx.Redis.Set(ctx, fmt.Sprintf("%s%d", redis.CommentReplyCountKey, commentID), fmt.Sprintf("%d", total), ttl)
}
