package types

type PageReq struct {
	Page     int64 `form:"page,default=1" validate:"min=1"`
	PageSize int64 `form:"pageSize,default=20" validate:"min=1,max=100"`
}

type PostContentItem struct {
	Type    int32  `json:"type"`
	Content string `json:"content"`
	Sort    int32  `json:"sort"`
}

type CreatePostReq struct {
	Title   string   `json:"title" validate:"required"`
	Cover   string   `json:"cover"`
	Type    int32    `json:"type" validate:"required"`
	Content string   `json:"content" validate:"required"`
	Topics  []string `json:"topics"`
}

type CreatePostResp struct {
	Code   int    `json:"code"`
	Msg    string `json:"msg"`
	PostId uint64 `json:"postId"`
}

type UpdatePostReq struct {
	PostId  uint64   `json:"postId" validate:"required"`
	Title   string   `json:"title" validate:"required"`
	Cover   string   `json:"cover"`
	Type    int32    `json:"type" validate:"required"`
	Content string   `json:"content" validate:"required"`
	Topics  []string `json:"topics"`
}

type UpdatePostResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type DeletePostReq struct {
	PostId uint64 `path:"postId" validate:"required"`
}

type DeletePostResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type PostTopic struct {
	PostId    uint64 `json:"postId"`
	TopicId   uint64 `json:"topicId"`
	TopicName string `json:"topicName"`
}

type PostInfo struct {
	Id              uint64 `json:"id"`
	UserId          uint64 `json:"userId"`
	Title           string `json:"title"`
	Cover           string `json:"cover"`
	CommentCount    uint64 `json:"commentCount"`
	CollectionCount uint64 `json:"collectionCount"`
	UpvoteCount     uint64 `json:"upvoteCount"`
	ShareCount      uint64 `json:"shareCount"`
	Visibility      int32  `json:"visibility"`
	IsTop           int32  `json:"isTop"`
	IsEssence       int32  `json:"isEssence"`
	IsLock          int32  `json:"isLock"`
	LatestRepliedAt int64  `json:"latestRepliedAt"`
	Tags            string `json:"tags"`
	CreatedAt       int64  `json:"createdAt"`
	UpdatedAt       int64  `json:"updatedAt"`
}

type PostDetailResp struct {
	Code    int         `json:"code"`
	Msg     string      `json:"msg"`
	Post    PostInfo    `json:"post"`
	Content string      `json:"content"`
	Topics  []PostTopic `json:"topics"`
}

type PostListResp struct {
	Code  int        `json:"code"`
	Msg   string     `json:"msg"`
	Total int64      `json:"total"`
	List  []PostInfo `json:"list"`
}

type GetPostListReq struct {
	UserId uint64 `form:"userId,optional"`
	PageReq
}

type GetPostDetailReq struct {
	PostId uint64 `path:"postId" validate:"required"`
}

type BatchGetPostReq struct {
	PostIds []uint64 `json:"postIds" validate:"required,min=1,max=100"`
}

type BatchGetPostResp struct {
	Code  int        `json:"code"`
	Msg   string     `json:"msg"`
	Posts []PostInfo `json:"posts"`
}

type IncViewCountReq struct {
	PostId uint64 `path:"postId" validate:"required"`
}

type IncViewCountResp struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}
