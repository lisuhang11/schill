package authctx

import (
	"context"
	"net/http"
	"strings"

	jwtx "SChill/common/jwt"

	"github.com/zeromicro/go-zero/rest"
)

func OptionalJWTMiddleware(secret string) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if secret == "" {
				next(w, r)
				return
			}

			token := extractBearerToken(r.Header.Get("Authorization"))
			if token == "" {
				next(w, r)
				return
			}

		claims, err := jwtx.ParseAccessToken(token, secret)
		if err != nil || claims == nil || claims.UserId == 0 {
			next(w, r)
			return
		}

		ctx := context.WithValue(r.Context(), userIDContextKey, claims.UserId)
		ctx = context.WithValue(ctx, tokenContextKey, token)
		next(w, r.WithContext(ctx))
		}
	}
}

func extractBearerToken(authHeader string) string {
	authHeader = strings.TrimSpace(authHeader)
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}
