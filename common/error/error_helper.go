package error

import (
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"strconv"
	"strings"
)

// RpcError 使用给定的业务错误码和消息构造一个 gRPC 错误。
// 该错误会被 status.Error 包装，其中 code 会被转换为 gRPC 状态码。
// 参数:
//   - code: 业务错误码（通常对应 gRPC codes.Code 类型）
//   - msg:  错误详细信息
//
// 返回:
//   - error: gRPC 状态错误
func RpcError(code int, msg string) error {
	return status.Error(codes.Code(code), msg)
}

// RpcBusinessError 使用预定义的消息映射构造业务错误。
// 通过 GetCodeMessage 获取对应错误码的默认消息，再构造 gRPC 错误。
// 参数:
//   - code: 业务错误码
//
// 返回:
//   - error: 包含默认消息的 gRPC 状态错误
func RpcBusinessError(code int) error {
	msg := GetCodeMessage(code)
	return status.Error(codes.Code(code), msg)
}

// ParseRpcError 解析 gRPC 错误，提取业务错误码和消息。
// 如果 err 为 nil，返回成功码及对应消息。
// 如果无法转换为 gRPC 状态错误，则返回内部错误码及对应消息。
// 如果错误码在业务范围（1000以上）内，直接返回该码和消息；
// 否则统一映射为内部错误码，并返回其默认消息（或原始消息）。
// 参数:
//   - err: 原始 error 对象
//
// 返回:
//   - int:   业务错误码
//   - string: 错误消息
func ParseRpcError(err error) (int, string) {
	if err == nil {
		return Success, GetCodeMessage(Success)
	}
	st, ok := status.FromError(err)
	if !ok {
		return ErrInternalError, GetCodeMessage(ErrInternalError)
	}
	code := int(st.Code())
	msg := st.Message()

	if code >= 1000 {
		return code, msg
	}

	if msg == "" {
		msg = GetCodeMessage(ErrInternalError)
	}
	return ErrInternalError, msg
}

// FormatRpcError 将错误码和消息格式化为字符串，格式为 "code:msg"。
// 该格式通常用于日志记录或跨组件传递时保持信息完整。
// 参数:
//   - code: 业务错误码
//   - msg:  错误消息
//
// 返回:
//   - string: 格式化后的错误字符串
func FormatRpcError(code int, msg string) string {
	return strconv.Itoa(code) + ":" + msg
}

// ParseFormattedError 解析由 FormatRpcError 生成的字符串，还原错误码和消息。
// 如果解析失败（如格式错误或无法转换为整数），则返回内部错误码和原始字符串。
// 参数:
//   - errStr: 格式化后的错误字符串（如 "1001:用户名已存在"）
//
// 返回:
//   - int:   业务错误码（解析失败时返回 ErrInternalError）
//   - string: 错误消息（解析失败时返回原始字符串）
func ParseFormattedError(errStr string) (int, string) {
	parts := strings.SplitN(errStr, ":", 2)
	if len(parts) == 2 {
		if code, err := strconv.Atoi(parts[0]); err == nil {
			return code, parts[1]
		}
	}
	return ErrInternalError, errStr
}
