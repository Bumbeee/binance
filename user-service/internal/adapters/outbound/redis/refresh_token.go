package redis

import (
	"context"
	"errors"
	"time"
	"userservice/internal/core/ports"

	"github.com/redis/go-redis/v9"
)

type RefreshTokenStore struct {
	client *redis.Client
}

func NewRefreshTokenStore(client *redis.Client) *RefreshTokenStore {
	return &RefreshTokenStore{
		client: client,
	}
}

func (s *RefreshTokenStore) Save(ctx context.Context, refreshToken, userID string, ttl time.Duration) error {
	return s.client.Set(ctx, refreshToken, userID, ttl).Err()
}

func (s *RefreshTokenStore) GetUserID(ctx context.Context, refreshToken string) (string, error) {
	userID, err := s.client.Get(ctx, refreshToken).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ports.ErrKeyNotFound
		}
		return "", err
	}
	return userID, nil
}

func (s *RefreshTokenStore) Delete(ctx context.Context, refreshToken string) error {
	return s.client.Del(ctx, refreshToken).Err()
}
