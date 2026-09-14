package interceptor

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	tools "market/shared/tools/grpc"
	"userservice/internal/core/domain"
	"userservice/internal/core/services/token"
)

type contextKey string

const (
	userIDContextKey   contextKey = "user_id"
	userRoleContextKey contextKey = "user_role"
)

var publicMethods = map[string]bool{
	"/userservice.v1.UserService/Register":      true,
	"/userservice.v1.UserService/Login":         true,
	"/userservice.v1.UserService/RefreshToken":  true,
	"/userservice.v1.UserService/ValidateToken": true, // используется другими сервисами
}

var requiredRoles = map[string]domain.Role{
	"/userservice.v1.UserService/GetUserProfile": domain.RoleAdmin,
}

type AuthInterceptor struct {
	validateToken *token.ValidateTokenCase
}

func NewAuthInterceptor(validateToken *token.ValidateTokenCase) *AuthInterceptor {
	return &AuthInterceptor{validateToken: validateToken}
}

func (a *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		accessToken, err := tools.ExtractToken(ctx)
		if err != nil {
			return nil, err
		}

		result := a.validateToken.Execute(ctx, accessToken)
		if !result.Valid {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		if requiredRole, needsCheck := requiredRoles[info.FullMethod]; needsCheck {
			if result.Role != requiredRole {
				return nil, status.Error(codes.PermissionDenied, "insufficient permissions")
			}
		}

		ctx = context.WithValue(ctx, userIDContextKey, result.UserID)
		ctx = context.WithValue(ctx, userRoleContextKey, string(result.Role))
		return handler(ctx, req)
	}
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	return tools.ValueFromContext(ctx, userIDContextKey)
}

func UserRoleFromContext(ctx context.Context) (domain.Role, bool) {
	role, ok := tools.ValueFromContext(ctx, userRoleContextKey)
	return domain.Role(role), ok
}
