package grpc

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"order-service/internal/core/domain"
)

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, domain.ErrOrderNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrNotOrderOwner):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, domain.ErrOrderAlreadyCancelled):
		return status.Error(codes.FailedPrecondition, err.Error())
	case errors.Is(err, domain.ErrInvalidSide),
		errors.Is(err, domain.ErrInvalidType),
		errors.Is(err, domain.ErrEmptyQuantity),
		errors.Is(err, domain.ErrPriceRequiredForLimit),
		errors.Is(err, domain.ErrPriceNotAllowedForMarket):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
