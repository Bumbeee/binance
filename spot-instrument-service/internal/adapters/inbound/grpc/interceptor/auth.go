package interceptor

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"spot-instrument-service/internal/core/ports"

	tools "market/shared/tools/grpc"
)

type contextKey string

const (
	userIDContextKey   contextKey = "user_id"
	userRoleContextKey contextKey = "user_role"
)

var publicMethods = map[string]bool{
	"/spotinstrumentservice.v1.SpotInstrumentService/GetInstrument":   true,
	"/spotinstrumentservice.v1.SpotInstrumentService/ListInstruments": true,
}

var requiredRoles = map[string]string{
	"/spotinstrumentservice.v1.SpotInstrumentService/CreateInstrument":  "ROLE_ADMIN",
	"/spotinstrumentservice.v1.SpotInstrumentService/ArchiveInstrument": "ROLE_ADMIN",
	"/spotinstrumentservice.v1.SpotInstrumentService/UpdateRate":        "ROLE_ADMIN",
}

type AuthInterceptor struct {
	validator ports.TokenValidator
}

func NewAuthInterceptor(validator ports.TokenValidator) *AuthInterceptor {
	return &AuthInterceptor{validator: validator}
}

func (a *AuthInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

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

		if requiredRole, needsCheck := requiredRoles[info.FullMethod]; needsCheck {
			if role != requiredRole {
				return nil, status.Error(codes.PermissionDenied, "insufficient permissions")
			}
		}

		ctx = context.WithValue(ctx, userIDContextKey, userID)
		ctx = context.WithValue(ctx, userRoleContextKey, role)
		return handler(ctx, req)
	}
}
