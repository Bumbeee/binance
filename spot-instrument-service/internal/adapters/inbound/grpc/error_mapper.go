package grpc

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"spot-instrument-service/internal/core/domain"
)

func toGRPCError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInstrumentAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrInstrumentNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrEmptySymbol),
		errors.Is(err, domain.ErrEmptyBaseAsset),
		errors.Is(err, domain.ErrEmptyQuoteAsset),
		errors.Is(err, domain.ErrInvalidPricePrecision),
		errors.Is(err, domain.ErrInvalidQuantityPrecision),
		errors.Is(err, domain.ErrEmptyMinOrderSize),
		errors.Is(err, domain.ErrEmptyRate):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
