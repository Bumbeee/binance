package grpc

import (
	user "market/proto/userservice/v1"
	"userservice/internal/core/services/auth"
	"userservice/internal/core/services/profile"
	"userservice/internal/core/services/token"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func toRegisterResponse(r *auth.RegisterResult) *user.RegisterResponse {
	return &user.RegisterResponse{
		UserId: r.UserID,
		AuthToken: &user.AuthToken{
			AccessToken:  r.AccessToken,
			RefreshToken: r.RefreshToken,
			ExpiresAt:    timestamppb.New(r.ExpiresAt),
		},
	}
}

func toLoginResponse(l *auth.LoginResult) *user.LoginResponse {
	return &user.LoginResponse{
		UserId: l.UserID,
		AuthToken: &user.AuthToken{
			AccessToken:  l.AccessToken,
			RefreshToken: l.RefreshToken,
			ExpiresAt:    timestamppb.New(l.ExpiresAt),
		},
	}
}

func toValidateTokenResponse(vt *token.ValidateTokenResult) *user.ValidateTokenResponse {
	return &user.ValidateTokenResponse{
		Valid:  vt.Valid,
		UserId: vt.UserID,
	}
}

func toRefreshTokenResponse(r *token.RefreshTokenResult) *user.RefreshTokenResponse {
	return &user.RefreshTokenResponse{
		AuthToken: &user.AuthToken{
			AccessToken:  r.AccessToken,
			RefreshToken: r.RefreshToken,
			ExpiresAt:    timestamppb.New(r.ExpiresAt),
		},
	}
}

func toGetProfileResponse(r profile.GetProfileResult) *user.GetProfileResponse {
	return &user.GetProfileResponse{
		UserId:    r.UserID,
		Email:     r.Email,
		Role:      toProtoRole(r.Role),
		CreatedAt: timestamppb.New(r.CreatedAt),
	}
}

func toProtoRole(r string) user.Role {
	switch r {
	case "admin":
		return user.Role_ROLE_ADMIN
	case "user":
		return user.Role_ROLE_USER
	default:
		return user.Role_ROLE_UNSPECIFIED
	}
}
