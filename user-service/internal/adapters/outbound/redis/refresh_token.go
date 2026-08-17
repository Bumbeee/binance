package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
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

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (s *RefreshTokenStore) Save(ctx context.Context, refreshToken, userID string, ttl time.Duration) error {
	return s.client.Set(ctx, hashToken(refreshToken), userID, ttl).Err()
}

func (s *RefreshTokenStore) GetUserID(ctx context.Context, refreshToken string) (string, error) {
	userID, err := s.client.Get(ctx, hashToken(refreshToken)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ports.ErrKeyNotFound
		}
		return "", err
	}
	return userID, nil
}

func (s *RefreshTokenStore) Delete(ctx context.Context, refreshToken string) error {
	return s.client.Del(ctx, hashToken(refreshToken)).Err()
}
