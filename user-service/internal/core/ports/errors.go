package ports

import "errors"

var (
	ErrKeyNotFound    = errors.New("key not found")
	ErrTokenExpired   = errors.New("token is expired")
	ErrMalformedToken = errors.New("token is malformed or has an invalid signature")
)
