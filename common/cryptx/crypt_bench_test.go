package cryptx

import (
	"testing"
)

// BenchmarkPasswordEncrypt 测试密码哈希性能
func BenchmarkPasswordEncrypt(b *testing.B) {
	password := "TestPassword123!"
	b.ReportAllocs()
	for b.Loop() {
		PasswordEncrypt(password)
	}
}

// BenchmarkPasswordEncryptLong 测试长密码哈希性能
func BenchmarkPasswordEncryptLong(b *testing.B) {
	password := "ThisIsAVeryLongPasswordWithMixedCharacters1234567890!@#$%^&*()"
	b.ReportAllocs()
	for b.Loop() {
		PasswordEncrypt(password)
	}
}

// BenchmarkPasswordVerify 测试密码验证性能
func BenchmarkPasswordVerify(b *testing.B) {
	password := "TestPassword123!"
	hashed, err := PasswordEncrypt(password)
	if err != nil {
		b.Fatal("encrypt failed:", err)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		if !PasswordVerify(hashed, password) {
			b.Fatal("verify failed")
		}
	}
}
