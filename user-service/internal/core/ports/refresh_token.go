package ports

import (
	"context"
	"errors"
	"time"
)

var ErrKeyNotFound = errors.New("key not found")

type RefreshTokenStore interface {
	Save(ctx context.Context, refreshToken, userID string, ttl time.Duration) error
	GetAndDelete(ctx context.Context, refreshToken string) (userID string, err error)
	Delete(ctx context.Context, refreshToken string) error
	DeleteAllByUserID(ctx context.Context, userID string) error
}
