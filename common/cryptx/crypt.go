package cryptx

import (
	"crypto/sha256"
	"fmt"
)

// PasswordEncrypt 直接对密码进行 SHA256 哈希
func PasswordEncrypt(password string) string {
	hash := sha256.Sum256([]byte(password))
	return fmt.Sprintf("%x", hash[:])
}

// PasswordVerify 验证明文密码是否与哈希匹配
func PasswordVerify(hashedPassword, plainPassword string) bool {
	return PasswordEncrypt(plainPassword) == hashedPassword
}
