package interceptor

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

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

		accessToken, err := extractToken(ctx)
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
		ctx = context.WithValue(ctx, userRoleContextKey, result.Role)
		return handler(ctx, req)
	}
}

func extractToken(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", status.Error(codes.Unauthenticated, "missing metadata")
	}
	values := md.Get("authorization")
	if len(values) == 0 {
		return "", status.Error(codes.Unauthenticated, "missing authorization header")
	}
	parts := strings.SplitN(values[0], " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return "", status.Error(codes.Unauthenticated, "invalid authorization header format")
	}
	return parts[1], nil
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDContextKey).(string)
	return userID, ok
}

func UserRoleFromContext(ctx context.Context) (domain.Role, bool) {
	role, ok := ctx.Value(userRoleContextKey).(domain.Role)
	return role, ok
}
