package model

type ESUser struct {
	ID              uint64 `json:"id"`
	Username        string `json:"username"`
	Status          int8   `json:"status"`
	IsAdmin         bool   `json:"is_admin"`
	CreatedAt       int64  `json:"created_at"`
	PostCount       int32  `json:"post_count"`
	CommentCount    int32  `json:"comment_count"`
	FollowerCount   int32  `json:"follower_count"`
	LikeCount       int32  `json:"like_count"`
	CollectionCount int32  `json:"collection_count"`
	LastActiveTime  int64  `json:"last_active_time"`
}

type ESPost struct {
	ID              uint64   `json:"id"`
	UserID          uint64   `json:"user_id"`
	Title           string   `json:"title"`
	Cover           string   `json:"cover"`
	Summary         string   `json:"summary"`
	Content         string   `json:"content"`
	UpdatedAt       int64    `json:"updated_at"`
	CommentCount    int64    `json:"comment_count"`
	CollectionCount int64    `json:"collection_count"`
	UpvoteCount     int64    `json:"upvote_count"`
	ShareCount      int64    `json:"share_count"`
	Visibility      int32    `json:"visibility"`
	IsTop           bool     `json:"is_top"`
	IsEssence       bool     `json:"is_essence"`
	IsLock          bool     `json:"is_lock"`
	Tags            []string `json:"tags"`
	CreatedAt       int64    `json:"created_at"`
	Username        string   `json:"username"`
	Avatar          string   `json:"avatar"`
	TopicNames      []string `json:"topic_names"`
}

type ESTopic struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	QuoteNum  int64  `json:"quote_num"`
	CreatedAt int64  `json:"created_at"`
}
