package jwt

import (
	"testing"
)

const (
	testSecret      = "test-secret-key-for-benchmark-12345678"
	testAccessExpire  = int64(3600)
	testRefreshExpire = int64(86400)
)

var benchUserID uint64 = 1234567890

// BenchmarkGenerateAccessToken 测试 Access Token 生成性能
func BenchmarkGenerateAccessToken(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, err := GenerateAccessToken(testAccessExpire, testSecret, benchUserID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerateRefreshToken 测试 Refresh Token 生成性能
func BenchmarkGenerateRefreshToken(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, err := GenerateRefreshToken(testRefreshExpire, testSecret, benchUserID)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkGenerateAccessTokenWithVersion 测试带版本号的 Access Token 生成性能
func BenchmarkGenerateAccessTokenWithVersion(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_, err := GenerateAccessTokenWithVersion(testAccessExpire, testSecret, benchUserID, 1)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseToken 测试 Token 解析性能
func BenchmarkParseToken(b *testing.B) {
	token, err := GenerateAccessToken(testAccessExpire, testSecret, benchUserID)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, err := ParseToken(token, testSecret)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkParseTokenUserID 测试 Token 解析 + 提取 UserID 性能
func BenchmarkParseTokenUserID(b *testing.B) {
	token, err := GenerateAccessToken(testAccessExpire, testSecret, benchUserID)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		_, err := ParseTokenUserID(token, testSecret)
		if err != nil {
			b.Fatal(err)
		}
	}
}
