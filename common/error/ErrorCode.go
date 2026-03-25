package error

const (
	// 通用错误
	Success    = 0   // 成功
	ErrUnknown = 999 // 未知错误

	// 用户模块错误码 (1000-1999)
	ErrUsernameOrPasswordEmpty = 1001 // 用户名或密码为空
	ErrUsernameExists          = 1002 // 用户名已存在
	ErrInvalidCredentials      = 1003 // 用户名或密码错误
	ErrAccountAbnormal         = 1004 // 账号异常（禁言/冻结）
	ErrInternalError           = 1005 // 内部错误
	ErrUnauthorized            = 1006 // 未授权访问
	ErrInvalidParams           = 1007 // 参数错误
	ErrUserNotExist            = 1008 // 用户不存在
	ErrInvalidRefreshToken     = 1009 // 无效的refresh token

	// 关系模块错误码 (2000-2999)
	ErrCannotFollowSelf = 2001 // 不能关注自己
	ErrAlreadyFollowed  = 2002 // 已经关注了
	ErrNotFollowing     = 2003 // 未关注

	// 内容模块错误码 (3000-3999)
	ErrPostNotExist     = 3001 // 帖子不存在
	ErrNoPermission     = 3002 // 没有操作权限
	ErrPostTitleEmpty   = 3003 // 帖子标题为空
	ErrPostContentEmpty = 3004 // 帖子内容为空

	// 评论模块错误码 (4000-4999)
	ErrCommentNotExist     = 4001 // 评论不存在
	ErrCommentContentEmpty = 4002 // 评论内容为空
	ErrAlreadyLiked        = 4003 // 已经点赞了
	ErrNotLiked            = 4004 // 未点赞
	// 可根据需要继续添加
)

var messageMap = map[int]string{
	Success:                    "ok",
	ErrUnknown:                 "未知错误",
	ErrUsernameOrPasswordEmpty: "用户名或密码为空",
	ErrUsernameExists:          "用户名已存在",
	ErrInvalidCredentials:      "用户名或密码错误",
	ErrAccountAbnormal:         "账号异常，无法登录",
	ErrInternalError:           "内部错误，请稍后重试",
	ErrUnauthorized:            "未授权访问，请先登录",
	ErrInvalidParams:           "参数错误",
	ErrUserNotExist:            "用户不存在",
	ErrInvalidRefreshToken:     "无效的刷新令牌",
	ErrCannotFollowSelf:        "不能关注自己",
	ErrAlreadyFollowed:         "已经关注了",
	ErrNotFollowing:            "未关注",
	ErrPostNotExist:            "帖子不存在",
	ErrNoPermission:            "没有操作权限",
	ErrPostTitleEmpty:          "帖子标题不能为空",
	ErrPostContentEmpty:        "帖子内容不能为空",
	ErrCommentNotExist:         "评论不存在",
	ErrCommentContentEmpty:     "评论内容不能为空",
	ErrAlreadyLiked:            "已经点赞了",
	ErrNotLiked:                "未点赞",
}

// GetCodeMessage 获取错误码对应的消息
func GetCodeMessage(code int) string {
	if msg, ok := messageMap[code]; ok {
		return msg
	}
	return "未知错误"
}
