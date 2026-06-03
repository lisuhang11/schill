package cryptx

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordEncrypt 使用 bcrypt 对密码进行哈希
// cost 设置为 12，在安全性和性能之间取得平衡
func PasswordEncrypt(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// PasswordVerify 验证明文密码是否与 bcrypt 哈希匹配
func PasswordVerify(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}
