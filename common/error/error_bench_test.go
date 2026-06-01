package error

import (
	"errors"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// BenchmarkRpcBusinessError 测试业务错误构造性能
func BenchmarkRpcBusinessError(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		RpcBusinessError(ErrInvalidCredentials)
	}
}

// BenchmarkParseRpcError 测试 gRPC 错误解析性能
func BenchmarkParseRpcError(b *testing.B) {
	err := RpcBusinessError(ErrInvalidCredentials)
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		code, msg := ParseRpcError(err)
		if code != ErrInvalidCredentials {
			b.Fatalf("expected code %d, got %d", ErrInvalidCredentials, code)
		}
		_ = msg
	}
}

// BenchmarkParseRpcError_Nil 测试 nil 错误解析性能
func BenchmarkParseRpcError_Nil(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		code, msg := ParseRpcError(nil)
		if code != Success {
			b.Fatalf("expected code %d, got %d", Success, code)
		}
		_ = msg
	}
}

// BenchmarkParseRpcError_NonGRPC 测试非 gRPC 错误解析性能
func BenchmarkParseRpcError_NonGRPC(b *testing.B) {
	err := errors.New("plain error")
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		code, _ := ParseRpcError(err)
		if code != ErrInternalError {
			b.Fatalf("expected code %d, got %d", ErrInternalError, code)
		}
	}
}

// BenchmarkFormatRpcError 测试错误格式化性能
func BenchmarkFormatRpcError(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		FormatRpcError(ErrInvalidCredentials, "用户名或密码错误")
	}
}

// BenchmarkParseFormattedError 测试格式化错误解析性能
func BenchmarkParseFormattedError(b *testing.B) {
	errStr := "1003:用户名或密码错误"
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		code, msg := ParseFormattedError(errStr)
		if code != ErrInvalidCredentials {
			b.Fatalf("expected code %d, got %d", ErrInvalidCredentials, code)
		}
		_ = msg
	}
}

// BenchmarkGetCodeMessage 测试错误码消息查找性能
func BenchmarkGetCodeMessage(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		msg := GetCodeMessage(ErrInvalidCredentials)
		if msg == "" {
			b.Fatal("empty message")
		}
	}
}

// BenchmarkGetCodeMessage_Unknown 测试未知错误码消息查找性能
func BenchmarkGetCodeMessage_Unknown(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		msg := GetCodeMessage(99999)
		_ = msg
	}
}

// BenchmarkStatusFromError 测试 gRPC status.FromError 性能
func BenchmarkStatusFromError(b *testing.B) {
	err := status.Error(codes.Code(ErrInvalidCredentials), "用户名或密码错误")
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		st, ok := status.FromError(err)
		if !ok {
			b.Fatal("status.FromError failed")
		}
		_ = st
	}
}
