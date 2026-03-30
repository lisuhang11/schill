package mq

// CommentCreatedMessage 评论创建消息
type CommentCreatedMessage struct {
	PostID    uint64 `json:"post_id"`
	CommentID uint64 `json:"comment_id"`
	UserID    uint64 `json:"user_id"`
}

// CommentDeletedMessage 评论删除消息
type CommentDeletedMessage struct {
	PostID    uint64 `json:"post_id"`
	CommentID uint64 `json:"comment_id"`
	UserID    uint64 `json:"user_id"`
}
