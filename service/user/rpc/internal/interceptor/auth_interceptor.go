package interceptor

import (
	"context"
	"strings"

	jwtx "SChill/common/jwt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthUnaryInterceptor returns a gRPC unary server interceptor that:
// - Extracts the JWT Bearer token from incoming gRPC metadata.
// - Parses the token to extract the authenticated userId.
// - Injects the userId into the Go context.
//
// The interceptor is optional: if no token is present, the request proceeds without a userId.
// It is the responsibility of individual RPC handlers to enforce authentication where required.
func AuthUnaryInterceptor(accessSecret string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if accessSecret == "" {
			return handler(ctx, req)
		}

		token := extractBearerFromMetadata(ctx)
		if token == "" {
			return handler(ctx, req)
		}

		claims, err := jwtx.ParseAccessToken(token, accessSecret)
		if err != nil || claims == nil || claims.UserId == 0 {
			return nil, status.Error(codes.Unauthenticated, "invalid access token")
		}

		ctx = context.WithValue(ctx, "userId", claims.UserId)
		return handler(ctx, req)
	}
}

func extractBearerFromMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values := md.Get("authorization")
	if len(values) == 0 {
		return ""
	}

	authHeader := strings.TrimSpace(values[0])
	if authHeader == "" {
		return ""
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}

	return strings.TrimSpace(parts[1])
}
