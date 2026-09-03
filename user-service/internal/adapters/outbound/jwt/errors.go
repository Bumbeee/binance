package jwt

import "errors"

var (
	ErrWrongSignMethod     = errors.New("unexpected signing method")
	ErrInvalidToken        = errors.New("invalid token")
	ErrInvalidSubjectClaim = errors.New("invalid subject claim")
)
