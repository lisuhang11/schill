package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

// Claims 自定义Claims结构，与go-zero兼容
type Claims struct {
	UserId       uint64 `json:"userId"`
	TokenVersion int64  `json:"tokenVersion,omitempty"`
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

// GenerateRefreshTokenWithVersion 生成带版本号的 Refresh Token（用于版本校验/撤销）
func GenerateRefreshTokenWithVersion(refreshExpire int64, refreshSecret string, userId uint64, version int64) (string, error) {
	now := time.Now()
	expire := now.Add(time.Duration(refreshExpire) * time.Second)
	claims := Claims{
		UserId:       userId,
		TokenVersion: version,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expire),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(refreshSecret))
}

// GenerateAccessTokenWithVersion 生成带版本号的 Access Token
func GenerateAccessTokenWithVersion(accessExpire int64, accessSecret string, userId uint64, version int64) (string, error) {
	now := time.Now()
	expire := now.Add(time.Duration(accessExpire) * time.Second)
	claims := Claims{
		UserId:       userId,
		TokenVersion: version,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expire),
			IssuedAt:  jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(accessSecret))
}

// ParseToken 解析 token 获取 Claims
func ParseToken(tokenString string, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// ParseTokenUserID 解析 token 获取用户 ID（保持向后兼容）
func ParseTokenUserID(tokenString string, secret string) (uint64, error) {
	claims, err := ParseToken(tokenString, secret)
	if err != nil {
		return 0, err
	}
	return claims.UserId, nil
}

// ParseRefreshToken 解析 refresh token 获取 Claims（含版本号）
func ParseRefreshToken(tokenString string, secret string) (*Claims, error) {
	return ParseToken(tokenString, secret)
}

// ParseAccessToken 解析 access token 获取 Claims（含版本号）
func ParseAccessToken(tokenString string, secret string) (*Claims, error) {
	return ParseToken(tokenString, secret)
}
