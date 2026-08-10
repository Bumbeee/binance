package grpc

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"userservice/internal/core/domain"
	"userservice/internal/core/services/auth"
	"userservice/internal/core/services/token"
)

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, domain.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())

	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.NotFound, err.Error())

	case errors.Is(err, domain.ErrInvalidEmail):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, domain.ErrEmptyPasswordHash):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, auth.ErrWeakPassword):
		return status.Error(codes.InvalidArgument, err.Error())

	case errors.Is(err, auth.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, err.Error())

	case errors.Is(err, token.ErrInvalidRefreshToken):
		return status.Error(codes.Unauthenticated, err.Error())

	default:
		return status.Error(codes.Internal, "internal error")
	}
}
