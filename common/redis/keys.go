package redis

const (
	KeyPrefix = "schill:"

	UserInfoKey         = KeyPrefix + "user:info:"
	UserProfileKey      = KeyPrefix + "user:profile:"
	UserStatKey         = KeyPrefix + "user:stat:"
	UserCacheVersionKey = KeyPrefix + "user:cache_version:"
	UserExpire          = 3600

	PostInfoKey         = KeyPrefix + "post:info:"
	PostContentKey      = KeyPrefix + "post:content:"
	PostDetailLockKey   = KeyPrefix + "post:detail:lock:"
	PostListKey         = KeyPrefix + "post:list:"
	PostCacheVersionKey = KeyPrefix + "post:cache_version:"
	PostListVersionKey  = KeyPrefix + "post:list:version:"
	PostExpire          = 1800
	CacheNullExpire     = 60

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
