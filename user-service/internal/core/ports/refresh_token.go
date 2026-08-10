package ports

import (
	"context"
	"errors"
	"time"
)

var ErrKeyNotFound = errors.New("key not found")

type RefreshTokenStore interface {
	Save(ctx context.Context, refreshToken, userID string, ttl time.Duration) error
	GetUserID(ctx context.Context, refreshToken string) (string, error)
	Delete(ctx context.Context, refreshToken string) error
}
