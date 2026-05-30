package mq

// CommentCreatedMessage 评论创建消息
type CommentCreatedMessage struct {
	PostID    uint64 `json:"post_id"`
	CommentID uint64 `json:"comment_id"`
	UserID    uint64 `json:"user_id"`
}

// CommentCreateEvent 评论创建事件消息，用于异步处理。
type CommentCreateEvent struct {
	CommentID     uint64 `json:"comment_id"`
	PostID        uint64 `json:"post_id"`
	UserID        uint64 `json:"user_id"`
	ParentID      uint64 `json:"parent_id"`
	ReplyToUserID uint64 `json:"reply_to_user_id"`
	Content       string `json:"content"`
	CreatedAt     int64  `json:"created_at"`
	Level         int32  `json:"level"`
}

// CommentDeletedMessage 评论删除消息
type CommentDeletedMessage struct {
	PostID    uint64 `json:"post_id"`
	CommentID uint64 `json:"comment_id"`
	UserID    uint64 `json:"user_id"`
}

// PostCollectMessage 帖子收藏消息
type PostCollectMessage struct {
	PostID    uint64 `json:"post_id"`
	UserID    uint64 `json:"user_id"`
	Timestamp int64  `json:"timestamp"`
}

// PostUncollectMessage 帖子取消收藏消息
type PostUncollectMessage struct {
	PostID    uint64 `json:"post_id"`
	UserID    uint64 `json:"user_id"`
	Timestamp int64  `json:"timestamp"`
}

// PostStarMessage 帖子点赞消息
type PostStarMessage struct {
	PostID    uint64 `json:"post_id"`
	UserID    uint64 `json:"user_id"`
	Timestamp int64  `json:"timestamp"`
}

// PostUnstarMessage 帖子取消点赞消息
type PostUnstarMessage struct {
	PostID    uint64 `json:"post_id"`
	UserID    uint64 `json:"user_id"`
	Timestamp int64  `json:"timestamp"`
}

// PostShareMessage 帖子分享消息
type PostShareMessage struct {
	PostID uint64 `json:"post_id"`
	UserID uint64 `json:"user_id"`
}

// UserFollowedMessage 用户关注消息
type UserFollowedMessage struct {
	FollowerID  uint64 `json:"follower_id"`
	FollowingID uint64 `json:"following_id"`
}

// UserUnfollowedMessage 用户取消关注消息
type UserUnfollowedMessage struct {
	FollowerID  uint64 `json:"follower_id"`
	FollowingID uint64 `json:"following_id"`
}

// UserMutualFollowMessage 用户双向关注消息
type UserMutualFollowMessage struct {
	UserID1 uint64 `json:"user_id_1"`
	UserID2 uint64 `json:"user_id_2"`
}

// VoteEvent 评论投票事件消息，用于异步落库。
type VoteEvent struct {
	CommentID uint64 `json:"comment_id"`
	UserID    uint64 `json:"user_id"`
	VoteType  int32  `json:"vote_type"` // 1:点赞 / 2:点踩 / 0:取消
	Timestamp int64  `json:"timestamp"`
}

// ContentChangedEvent 内容变更事件，供 feed / search 等下游消费。
type ContentChangedEvent struct {
	ContentID   uint64   `json:"content_id"`
	AuthorID    uint64   `json:"author_id"`
	OpType      string   `json:"op_type"`
	ContentType string   `json:"content_type"`
	Visibility  int32    `json:"visibility"`
	Status      string   `json:"status"`
	Title       string   `json:"title"`
	Summary     string   `json:"summary"`
	Cover       string   `json:"cover"`
	TopicNames  []string `json:"topic_names"`
	Tags        []string `json:"tags"`
	CreatedAt   int64    `json:"created_at"`
	UpdatedAt   int64    `json:"updated_at"`
}
