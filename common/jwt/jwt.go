package jwt

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"time"
)

// Claims 自定义Claims结构，与go-zero兼容
type Claims struct {
	UserId uint64 `json:"userId"`
	jwt.RegisteredClaims
}

// GenerateAccessToken 生成 Access Token
func GenerateAccessToken(accessExpire int64, accessSecret string, userId uint64) (string, error) {
	now := time.Now()
	expire := now.Add(time.Duration(accessExpire) * time.Second)
	claims := Claims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expire),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(accessSecret))
}

// GenerateRefreshToken 生成 Refresh Token
func GenerateRefreshToken(refreshExpire int64, refreshSecret string, userId uint64) (string, error) {
	now := time.Now()
	expire := now.Add(time.Duration(refreshExpire) * time.Second)
	claims := Claims{
		UserId: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expire),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(refreshSecret))
}

// ParseToken 解析 token 获取用户 ID
func ParseToken(tokenString string, secret string) (uint64, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return 0, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims.UserId, nil
	}

	return 0, errors.New("invalid token")
}

// ParseRefreshToken 解析 refresh token
func ParseRefreshToken(tokenString string, secret string) (uint64, error) {
	return ParseToken(tokenString, secret)
}

// ParseAccessToken 解析 access token
func ParseAccessToken(tokenString string, secret string) (uint64, error) {
	return ParseToken(tokenString, secret)
}
