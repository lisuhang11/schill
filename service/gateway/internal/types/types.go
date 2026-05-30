package types

type Response struct {
	Code int64       `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

type RegisterReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginReq struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RefreshTokenReq struct {
	RefreshToken string `json:"refreshToken"`
}

type UpdateProfileReq struct {
	Gender    *int64  `json:"gender,optional"`
	Birthday  *string `json:"birthday,optional"`
	Signature *string `json:"signature,optional"`
	Location  *string `json:"location,optional"`
	Website   *string `json:"website,optional"`
	Company   *string `json:"company,optional"`
	JobTitle  *string `json:"jobTitle,optional"`
	Education *string `json:"education,optional"`
}

type UpdateAvatarReq struct {
	AvatarUrl string `json:"avatarUrl"`
}

type CreatePostReq struct {
	Title      string            `json:"title"`
	Cover      string            `json:"cover,optional"`
	Visibility int32             `json:"visibility,default=0"`
	Contents   []PostContentItem `json:"contents"`
	Topics     []string          `json:"topics,optional"`
	Tags       string            `json:"tags,optional"`
}

type UpdatePostReq struct {
	Title      string            `json:"title"`
	Cover      string            `json:"cover,optional"`
	Visibility int32             `json:"visibility,default=0"`
	Contents   []PostContentItem `json:"contents"`
	Topics     []string          `json:"topics,optional"`
	Tags       string            `json:"tags,optional"`
}

type PostContentItem struct {
	Type    int32  `json:"type"`
	Content string `json:"content"`
	Sort    int32  `json:"sort"`
}

type CreateCommentReq struct {
	PostId        uint64 `json:"postId"`
	ParentId      uint64 `json:"parentId,optional"`
	ReplyToUserId uint64 `json:"replyToUserId,optional"`
	Content       string `json:"content"`
}

type VoteCommentReq struct {
	VoteType int32 `json:"voteType"`
}
