package domain

import "errors"

var (
	ErrEmptySymbol              = errors.New("symbol is empty")
	ErrEmptyBaseAsset           = errors.New("base asset is empty")
	ErrEmptyQuoteAsset          = errors.New("quote asset is empty")
	ErrInvalidPricePrecision    = errors.New("price precision should be positive")
	ErrInvalidQuantityPrecision = errors.New("quantity precision should be positive")
	ErrEmptyMinOrderSize        = errors.New("min order size is empty")
	ErrEmptyRate                = errors.New("rate is empty")
	ErrInstrumentNotFound       = errors.New("instrument not found")
	ErrInstrumentAlreadyExists  = errors.New("instrument already exists")
)
