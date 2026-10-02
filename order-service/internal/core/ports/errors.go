package ports

import "errors"

var (
	ErrDuplicateIdempotencyKey = errors.New("order with this idempotency key already exists")
	ErrNoTradesForInstrument   = errors.New("no trades have executed for this instrument yet")
)
