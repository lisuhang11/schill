package mqs

import (
	"context"
	"fmt"

	"SChill/common/redis"
	"SChill/service/content/rpc/internal/model"
	"SChill/service/content/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

// invalidatePostListForPost bumps the post list cache version for the post's author.
// This causes all list cache keys for that author to miss, forcing a DB rebuild
// with latest stats on the next request.
func invalidatePostListForPost(svcCtx *svc.ServiceContext, postID uint64) {
	var post model.Post
	if err := svcCtx.DB.Select("user_id").Where("id = ?", postID).First(&post).Error; err != nil {
		logx.Errorf("invalidatePostListForPost: query post author failed postId=%d err=%v", postID, err)
		return
	}
	if post.UserID > 0 {
		_, _ = svcCtx.Redis.Incr(context.Background(), fmt.Sprintf("%s%d", redis.PostListVersionKey, post.UserID))
	}
}
