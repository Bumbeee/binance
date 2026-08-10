package domain

import "errors"

var (
	ErrInvalidEmail      = errors.New("invalid email")
	ErrInvalidCreds      = errors.New("invalid credentials")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	ErrUserNotFound      = errors.New("there is no user with this email")
	ErrEmptyPasswordHash = errors.New("password hash is empty")
)
