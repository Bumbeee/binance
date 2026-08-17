package grpc

import (
	"context"
	"userservice/internal/adapters/inbound/grpc/interceptor"
	"userservice/internal/core/services/auth"
	"userservice/internal/core/services/profile"
	"userservice/internal/core/services/token"

	user "market/proto/userservice/v1"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	user.UnimplementedUserServiceServer
	register       *auth.RegisterCase
	login          *auth.LoginCase
	logout         *auth.LogoutCase
	validateToken  *token.ValidateTokenCase
	refreshToken   *token.RefreshTokenCase
	getProfile     *profile.GetProfileCase
	getUserProfile *profile.GetUserProfileCase
	changePassword *auth.ChangePasswordCase
}

func NewServer(
	register *auth.RegisterCase,
	login *auth.LoginCase,
	logout *auth.LogoutCase,
	validateToken *token.ValidateTokenCase,
	refreshToken *token.RefreshTokenCase,
	getProfile *profile.GetProfileCase,
	getUserProfile *profile.GetUserProfileCase,
	changePassword *auth.ChangePasswordCase,
) *Server {
	return &Server{
		register:       register,
		login:          login,
		logout:         logout,
		validateToken:  validateToken,
		refreshToken:   refreshToken,
		getProfile:     getProfile,
		getUserProfile: getUserProfile,
		changePassword: changePassword,
	}
}

func (s *Server) Register(ctx context.Context, req *user.RegisterRequest) (*user.RegisterResponse, error) {
	res, err := s.register.Execute(ctx, req.Email, req.Password)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toRegisterResponse(res), nil
}

func (s *Server) Login(ctx context.Context, req *user.LoginRequest) (*user.LoginResponse, error) {
	res, err := s.login.Execute(ctx, req.Email, req.Password)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toLoginResponse(res), nil
}

func (s *Server) Logout(ctx context.Context, req *user.LogoutRequest) (*user.LogoutResponse, error) {
	if err := s.logout.Execute(ctx, req.RefreshToken); err != nil {
		return nil, toGRPCError(err)
	}
	return &user.LogoutResponse{}, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *user.ValidateTokenRequest) (*user.ValidateTokenResponse, error) {
	res := s.validateToken.Execute(ctx, req.AccessToken)
	return toValidateTokenResponse(res), nil
}

func (s *Server) RefreshToken(ctx context.Context, req *user.RefreshTokenRequest) (*user.RefreshTokenResponse, error) {
	res, err := s.refreshToken.Execute(ctx, req.RefreshToken)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toRefreshTokenResponse(res), nil
}

func (s *Server) GetProfile(ctx context.Context, req *user.GetProfileRequest) (*user.GetProfileResponse, error) {
	userID, ok := interceptor.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "user id not found in context")
	}

	res, err := s.getProfile.Execute(ctx, userID)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return toGetProfileResponse(res), nil
}

func (s *Server) GetUserProfile(ctx context.Context, req *user.GetUserProfileRequest) (*user.GetUserProfileResponse, error) {
	res, err := s.getUserProfile.Execute(ctx, req.UserId)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toGetUserProfileResponse(res), nil
}

func (s *Server) ChangePassword(ctx context.Context, req *user.ChangePasswordRequest) (*user.ChangePasswordResponse, error) {
	userID, ok := interceptor.UserIDFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Internal, "user id not found in context")
	}

	if err := s.changePassword.Execute(ctx, userID, req.OldPassword, req.NewPassword); err != nil {
		return nil, toGRPCError(err)
	}
	return &user.ChangePasswordResponse{}, nil
}
