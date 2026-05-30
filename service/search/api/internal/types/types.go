package types

type PageReq struct {
	Page     int64 `form:"page,default=1" validate:"min=1"`
	PageSize int64 `form:"pageSize,default=20" validate:"min=1,max=100"`
}

type SearchPostReq struct {
	Keyword string `form:"keyword" validate:"required"`
	PageReq
}

type SearchPostResp struct {
	Code  int64            `json:"code"`
	Msg   string           `json:"msg"`
	Total int64            `json:"total"`
	List  []SearchPostItem `json:"list"`
}

type SearchPostItem struct {
	Id              int64    `json:"id"`
	UserId          int64    `json:"userId"`
	Title           string   `json:"title"`
	Cover           string   `json:"cover"`
	Summary         string   `json:"summary"`
	Username        string   `json:"username"`
	Avatar          string   `json:"avatar"`
	Content         string   `json:"content"`
	TopicNames      []string `json:"topicNames"`
	CommentCount    int64    `json:"commentCount"`
	CollectionCount int64    `json:"collectionCount"`
	UpvoteCount     int64    `json:"upvoteCount"`
	ShareCount      int64    `json:"shareCount"`
	Visibility      int8     `json:"visibility"`
	IsTop           bool     `json:"isTop"`
	IsEssence       bool     `json:"isEssence"`
	IsLock          bool     `json:"isLock"`
	Tags            []string `json:"tags"`
	CreatedAt       int64    `json:"createdAt"`
}

type SearchUserReq struct {
	Keyword string `form:"keyword" validate:"required"`
	PageReq
}

type SearchUserResp struct {
	Code  int64            `json:"code"`
	Msg   string           `json:"msg"`
	Total int64            `json:"total"`
	List  []SearchUserItem `json:"list"`
}

type SearchUserItem struct {
	Id              int64  `json:"id"`
	Username        string `json:"username"`
	Status          int8   `json:"status"`
	IsAdmin         bool   `json:"isAdmin"`
	PostCount       int32  `json:"postCount"`
	CommentCount    int32  `json:"commentCount"`
	FollowerCount   int32  `json:"followerCount"`
	LikeCount       int32  `json:"likeCount"`
	CollectionCount int32  `json:"collectionCount"`
	LastActiveTime  int64  `json:"lastActiveTime"`
	CreatedAt       int64  `json:"createdAt"`
}

type SearchTopicReq struct {
	Keyword string `form:"keyword" validate:"required"`
	PageReq
}

type SearchTopicResp struct {
	Code  int64             `json:"code"`
	Msg   string            `json:"msg"`
	Total int64             `json:"total"`
	List  []SearchTopicItem `json:"list"`
}

type SearchTopicItem struct {
	Id        int64  `json:"id"`
	Name      string `json:"name"`
	QuoteNum  int64  `json:"quoteNum"`
	CreatedAt int64  `json:"createdAt"`
}
