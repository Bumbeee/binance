package interceptor

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	tools "market/shared/tools/grpc"
	"order-service/internal/core/ports"
)

type contextKey string

const (
	userIDContextKey   contextKey = "user_id"
	userRoleContextKey contextKey = "user_role"
)

type AuthInterceptor struct {
	validator ports.TokenValidator
}

func NewAuthInterceptor(validator ports.TokenValidator) *AuthInterceptor {
	return &AuthInterceptor{validator: validator}
}

func (a *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		accessToken, err := tools.ExtractToken(ctx)
		if err != nil {
			return nil, err
		}

		userID, role, valid, err := a.validator.Validate(ctx, accessToken)
		if err != nil {
			return nil, status.Error(codes.Unavailable, "failed to validate token")
		}
		if !valid {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		ctx = context.WithValue(ctx, userIDContextKey, userID)
		ctx = context.WithValue(ctx, userRoleContextKey, role)
		return handler(ctx, req)
	}
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	return tools.ValueFromContext(ctx, userIDContextKey)
}

func UserRoleFromContext(ctx context.Context) (string, bool) {
	return tools.ValueFromContext(ctx, userRoleContextKey)
}
