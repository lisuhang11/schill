package redis

const (
	KeyPrefix = "schill:"

	UserInfoKey         = KeyPrefix + "user:info:"
	UserProfileKey      = KeyPrefix + "user:profile:"
	UserStatKey         = KeyPrefix + "user:stat:"
	UserCacheVersionKey = KeyPrefix + "user:cache_version:"
	UserExpire          = 3600

	// --- Post detail: layered caching ---
	// base: immutable fields (title, cover, content, topics, created_at, user_id, visibility, etc.)
	//       cached with long TTL, invalidated only on post edit/delete
	PostBaseKey = KeyPrefix + "post:base:"
	// stats: mutable counters (upvote_count, collection_count, comment_count, share_count)
	//        cached with short TTL or computed from interaction-rpc in real-time
	PostStatsKey = KeyPrefix + "post:stats:"
	// viewer_state: per-user state (is_starred, is_collected, is_following)
	//               NOT cached — always fetched from interaction/relation-rpc

	PostInfoKey         = KeyPrefix + "post:info:"   // legacy, replaced by PostBaseKey + PostStatsKey
	PostContentKey      = KeyPrefix + "post:content:" // legacy
	PostDetailLockKey   = KeyPrefix + "post:detail:lock:"
	PostListKey         = KeyPrefix + "post:list:"
	PostCacheVersionKey = KeyPrefix + "post:cache_version:"
	PostListVersionKey  = KeyPrefix + "post:list:version:"

	// PostBaseExpire is the long TTL for post base (title, cover, content, topics).
	// Invalidated only when post is edited or deleted.
	PostBaseExpire = 600 // 10 minutes
	// PostStatsExpire is the short TTL for post stats (counters).
	// Stats change frequently via user interactions.
	PostStatsExpire = 30 // 30 seconds
	// PostExpire is the TTL for post list cache.
	// Short TTL because stats (counters) are embedded and change frequently.
	PostExpire      = 300
	CacheNullExpire = 60

	// FeedCacheExpire is the TTL for feed aggregation cache (feed-rpc).
	// Short TTL because feeds are personalized and change frequently.
	FeedCacheExpire = 10 // 10 seconds
	// FeedFollowingCacheExpire is slightly longer for following feeds.
	FeedFollowingCacheExpire = 30 // 30 seconds
	// FeedMaxCachedPages is the max number of pages to cache per user+feedType.
	FeedMaxCachedPages = 3

	// --- Interaction (like/collection) Redis-as-source-of-truth ---
	InteractionDefaultTTL = 86400 * 7 // 7 days for like/collection sets and counters

	FollowStatusKey           = KeyPrefix + "follow:status:"
	FollowerListKey           = KeyPrefix + "follower:list:"
	FollowingListKey          = KeyPrefix + "following:list:"
	FollowExpire              = 600
	RelationFollowingZSetKey  = KeyPrefix + "relation:following:"
	RelationFollowersZSetKey  = KeyPrefix + "relation:followers:"
	RelationFollowingEmptyKey = KeyPrefix + "relation:following:empty:"
	RelationFollowersEmptyKey = KeyPrefix + "relation:followers:empty:"
	RelationCacheExpire       = 86400 * 7
	RelationEmptyExpire       = 60

	SearchResultKey = KeyPrefix + "search:result:"
	SearchExpire    = 300
	HotSearchKey    = KeyPrefix + "search:hot"

	PostLikeRelationKey = KeyPrefix + "like:entity:post:"
	PostLikeCountKey    = KeyPrefix + "like_count:entity:post:"
	UserLikesKey        = KeyPrefix + "user_likes:"

	PostCollectionRelationKey = KeyPrefix + "collect:entity:post:"
	PostCollectionCountKey    = KeyPrefix + "collect_count:entity:post:"
	UserCollectionsKey        = KeyPrefix + "user_collects:"

	// 评论相关
	CommentInfoKey        = KeyPrefix + "comment:info:"
	CommentContentKey     = KeyPrefix + "comment:content:"
	PostCommentsKey       = KeyPrefix + "post:comments:"   // + {postId}:list / :hot
	CommentRepliesKey     = KeyPrefix + "comment:replies:" // + {rootId}:list
	CommentRepliesMetaKey = KeyPrefix + "comment:replies:meta:"
	CommentVoteKey        = KeyPrefix + "comment:vote:"    // + {commentId}:user:{userId}
	UserVoteCountKey      = KeyPrefix + "user:vote:count:" // + {userId}:{date}
	PostCommentCountKey   = KeyPrefix + "post:comment_count:"
	PostCommentsMetaKey   = KeyPrefix + "post:comments:meta:"
	CommentReplyCountKey  = KeyPrefix + "comment:reply_count:"
	CommentLockKey        = KeyPrefix + "comment:lock:" // + {postId}
	CommentIdGenerator    = KeyPrefix + "comment:id:gen"
	CommentExpire         = 600
	VoteExpire            = 86400 * 30 // 30天
)
