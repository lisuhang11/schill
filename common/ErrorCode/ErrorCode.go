package errorcode

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
	ErrInvalidParams           = 1007
	ErrUserNotExist            = 1008
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
	ErrInvalidParams:           "ErrInvalidParams",
	ErrUserNotExist:            "用户不存在",
}

// GetCodeMessage 获取错误码对应的消息
func GetCodeMessage(code int) string {
	if msg, ok := messageMap[code]; ok {
		return msg
	}
	return "未知错误"
}
