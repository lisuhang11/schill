package jwt

import (
	"github.com/golang-jwt/jwt/v4"
	"time"
)

// GenerateAccessToken 生成 Access Token
func GenerateAccessToken(accessExpire int64, accessSecret string, userId uint64) (string, error) {
	now := time.Now()
	expire := now.Add(time.Duration(accessExpire) * time.Second)
	claims := jwt.MapClaims{
		"userId": userId,
		"iat":    now.Unix(),
		"exp":    expire.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(accessSecret))
}

// GenerateRefreshToken 生成 Refresh Token
func GenerateRefreshToken(refreshExpire int64, refreshSecret string, userId uint64) (string, error) {
	now := time.Now()
	expire := now.Add(time.Duration(refreshExpire) * time.Second)
	claims := jwt.MapClaims{
		"userId": userId,
		"iat":    now.Unix(),
		"exp":    expire.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(refreshSecret))
}
