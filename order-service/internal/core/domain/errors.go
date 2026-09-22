package domain

import "errors"

var (
	ErrInvalidSide              = errors.New("invalid order side")
	ErrInvalidType              = errors.New("invalid order type")
	ErrEmptyQuantity            = errors.New("quantity is empty")
	ErrPriceRequiredForLimit    = errors.New("price is required for limit orders")
	ErrPriceNotAllowedForMarket = errors.New("price is not allowed for market orders")
	ErrOrderNotFound            = errors.New("order not found")
	ErrOrderAlreadyTerminal     = errors.New("order is already in a terminal state")
	ErrNotOrderOwner            = errors.New("caller does not own this order")
)
