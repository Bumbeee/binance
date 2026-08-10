package interceptor

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"userservice/internal/core/services/token"
)

type contextKey string

const userIDContextKey contextKey = "user_id"

var publicMethods = map[string]bool{
	"/userservice.v1.UserService/Register":      true,
	"/userservice.v1.UserService/Login":         true,
	"/userservice.v1.UserService/RefreshToken":  true,
	"/userservice.v1.UserService/ValidateToken": true, // используется другими сервисами, у которых тоже ещё нет своего access_token
}

// Created with AI

type AuthInterceptor struct {
	validateToken *token.ValidateTokenCase
}

func NewAuthInterceptor(validateToken *token.ValidateTokenCase) *AuthInterceptor {
	return &AuthInterceptor{validateToken: validateToken}
}

func (a *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
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

		ctx = context.WithValue(ctx, userIDContextKey, result.UserID)
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
