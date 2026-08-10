package auth

import "errors"

var (
	ErrWeakPassword       = errors.New("password too weak")
	ErrInvalidCredentials = errors.New("invalid credentials")
)
