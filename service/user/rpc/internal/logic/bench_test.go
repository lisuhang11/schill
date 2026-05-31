package logic

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"SChill/common/cryptx"
	"SChill/common/jwt"
)

// =============================================================================
// 登录链路核心逻辑性能测试
// =============================================================================

// BenchmarkPasswordVerify 密码验证（登录链路第一步）
func BenchmarkPasswordVerify(b *testing.B) {
	password := "TestPass123!"
	hashed := cryptx.PasswordEncrypt(password)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		if !cryptx.PasswordVerify(hashed, password) {
			b.Fatal("verify failed")
		}
	}
}

// BenchmarkTokenGeneration 完整登录 Token 生成链路
func BenchmarkTokenGeneration(b *testing.B) {
	const accessExpire = int64(3600)
	const refreshExpire = int64(86400)
	const accessSecret = "test-access-secret-1234567890"
	const refreshSecret = "test-refresh-secret-1234567890"
	userID := uint64(12345)
	version := time.Now().Unix()

	b.ReportAllocs()
	for b.Loop() {
		accessToken, err := jwt.GenerateAccessTokenWithVersion(accessExpire, accessSecret, userID, version)
		if err != nil {
			b.Fatal(err)
		}
		refreshToken, err := jwt.GenerateRefreshTokenWithVersion(refreshExpire, refreshSecret, userID, version)
		if err != nil {
			b.Fatal(err)
		}
		if accessToken == "" || refreshToken == "" {
			b.Fatal("empty token")
		}
	}
}

// BenchmarkTokenValidation 完整 Token 验证链路
func BenchmarkTokenValidation(b *testing.B) {
	const accessExpire = int64(3600)
	const accessSecret = "test-access-secret-1234567890"
	userID := uint64(12345)

	token, _ := jwt.GenerateAccessTokenWithVersion(accessExpire, accessSecret, userID, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		claims, err := jwt.ParseToken(token, accessSecret)
		if err != nil {
			b.Fatal(err)
		}
		if claims.UserId != userID {
			b.Fatalf("expected userId %d, got %d", userID, claims.UserId)
		}
	}
}

// BenchmarkTokenRefresh 完整 Token 刷新链路
func BenchmarkTokenRefresh(b *testing.B) {
	const accessExpire = int64(3600)
	const refreshExpire = int64(86400)
	const accessSecret = "test-access-secret-1234567890"
	const refreshSecret = "test-refresh-secret-1234567890"
	userID := uint64(12345)

	oldRefreshToken, _ := jwt.GenerateRefreshTokenWithVersion(refreshExpire, refreshSecret, userID, 1)
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		claims, err := jwt.ParseRefreshToken(oldRefreshToken, refreshSecret)
		if err != nil {
			b.Fatal(err)
		}
		newAccessToken, err := jwt.GenerateAccessTokenWithVersion(accessExpire, accessSecret, claims.UserId, claims.TokenVersion)
		if err != nil {
			b.Fatal(err)
		}
		newRefreshToken, err := jwt.GenerateRefreshTokenWithVersion(refreshExpire, refreshSecret, claims.UserId, claims.TokenVersion)
		if err != nil {
			b.Fatal(err)
		}
		if newAccessToken == "" || newRefreshToken == "" {
			b.Fatal("empty token")
		}
	}
}

// =============================================================================
// 密码强度验证性能测试
// =============================================================================

// BenchmarkValidatePassword 密码强度验证
func BenchmarkValidatePassword(b *testing.B) {
	password := "StrongPass123"
	b.ReportAllocs()
	for b.Loop() {
		if len(password) < 8 {
			b.Fatal("too short")
		}
		var hasUpper, hasLower, hasDigit bool
		for _, ch := range password {
			switch {
			case ch >= 'A' && ch <= 'Z':
				hasUpper = true
			case ch >= 'a' && ch <= 'z':
				hasLower = true
			case ch >= '0' && ch <= '9':
				hasDigit = true
			}
		}
		if !hasUpper || !hasLower || !hasDigit {
			b.Fatal("password too weak")
		}
	}
}

// BenchmarkValidatePasswordLong 长密码强度验证
func BenchmarkValidatePasswordLong(b *testing.B) {
	password := "ThisIsAVeryLongPasswordWithAllRequiredChars123"
	b.ReportAllocs()
	for b.Loop() {
		if len(password) < 8 {
			b.Fatal("too short")
		}
		var hasUpper, hasLower, hasDigit bool
		for _, ch := range password {
			switch {
			case ch >= 'A' && ch <= 'Z':
				hasUpper = true
			case ch >= 'a' && ch <= 'z':
				hasLower = true
			case ch >= '0' && ch <= '9':
				hasDigit = true
			}
		}
		if !hasUpper || !hasLower || !hasDigit {
			b.Fatal("password too weak")
		}
	}
}

// =============================================================================
// 缓存 Key 构建性能测试
// =============================================================================

// BenchmarkCacheKeyGeneration 缓存 Key 生成
func BenchmarkCacheKeyGeneration(b *testing.B) {
	userID := uint64(12345)
	b.ReportAllocs()
	for b.Loop() {
		_ = buildUserInfoCacheKey(userID)
		_ = buildUserProfileCacheKey(userID)
		_ = buildUserStatCacheKey(userID)
		_ = buildUserBasicInfoCacheKey(userID)
		_ = tokenVersionKey(userID)
	}
}

// BenchmarkCacheKeyGenerationBatch 批量缓存 Key 生成
func BenchmarkCacheKeyGenerationBatch(b *testing.B) {
	userIDs := make([]uint64, 100)
	for i := range userIDs {
		userIDs[i] = uint64(1000 + i)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		for _, uid := range userIDs {
			_ = buildUserInfoCacheKey(uid)
		}
	}
}

// =============================================================================
// 并发性能测试 — 模拟高并发登录/Token 操作
// =============================================================================

// BenchmarkConcurrentPasswordVerify 并发密码验证
func BenchmarkConcurrentPasswordVerify(b *testing.B) {
	password := "TestPass123!"
	hashed := cryptx.PasswordEncrypt(password)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if !cryptx.PasswordVerify(hashed, password) {
				b.Fatal("verify failed")
			}
		}
	})
}

// BenchmarkConcurrentTokenGeneration 并发 Token 生成
func BenchmarkConcurrentTokenGeneration(b *testing.B) {
	const accessExpire = int64(3600)
	const accessSecret = "test-access-secret-1234567890"

	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		// Each goroutine has its own userID to avoid collision in realistic scenarios
		userID := uint64(10000 + b.N)
		version := time.Now().Unix()
		for pb.Next() {
			token, err := jwt.GenerateAccessTokenWithVersion(accessExpire, accessSecret, userID, version)
			if err != nil || token == "" {
				b.Fatal("token generation failed")
			}
		}
	})
}

// BenchmarkConcurrentTokenValidation 并发 Token 验证
func BenchmarkConcurrentTokenValidation(b *testing.B) {
	const accessExpire = int64(3600)
	const accessSecret = "test-access-secret-1234567890"

	token, _ := jwt.GenerateAccessTokenWithVersion(accessExpire, accessSecret, 12345, 1)
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, err := jwt.ParseToken(token, accessSecret)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkConcurrentCacheKeyGeneration 并发缓存 Key 生成
func BenchmarkConcurrentCacheKeyGeneration(b *testing.B) {
	b.ReportAllocs()
	b.RunParallel(func(pb *testing.PB) {
		userID := uint64(10000)
		for pb.Next() {
			_ = buildUserInfoCacheKey(userID)
			_ = tokenVersionKey(userID)
			userID++
		}
	})
}

// =============================================================================
// 字符串处理性能测试（常见于请求处理）
// =============================================================================

// BenchmarkUsernameValidation 用户名验证
func BenchmarkUsernameValidation(b *testing.B) {
	username := "testuser123"
	b.ReportAllocs()
	for b.Loop() {
		if len(username) < 3 || len(username) > 20 {
			b.Fatal("invalid username length")
		}
		if strings.ContainsAny(username, " \t\n\r") {
			b.Fatal("username contains whitespace")
		}
	}
}

// BenchmarkUserIDStringConversion 用户 ID 字符串转换
func BenchmarkUserIDStringConversion(b *testing.B) {
	userID := uint64(1234567890)
	b.ReportAllocs()
	for b.Loop() {
		s := fmt.Sprintf("%d", userID)
		if s == "" {
			b.Fatal("empty string")
		}
		_, err := parseUint(s)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func parseUint(s string) (uint64, error) {
	var n uint64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid char: %c", c)
		}
		n = n*10 + uint64(c-'0')
	}
	return n, nil
}

// =============================================================================
// Context 操作性能测试
// =============================================================================

// BenchmarkContextCreation Context 创建
func BenchmarkContextCreation(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_ = ctx
		cancel()
	}
}

// BenchmarkContextWithValue Context 带值传递
func BenchmarkContextWithValue(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		ctx := context.WithValue(context.Background(), "userId", uint64(12345))
		val := ctx.Value("userId")
		if val == nil {
			b.Fatal("nil value")
		}
	}
}
