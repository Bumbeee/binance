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
	hashedToken := hashToken(refreshToken)
	pipe := s.client.TxPipeline()
	pipe.Set(ctx, hashedToken, userID, ttl)
	pipe.SAdd(ctx, "user_tokens:"+userID, hashedToken)
	pipe.Expire(ctx, "user_tokens:"+userID, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func (s *RefreshTokenStore) GetAndDelete(ctx context.Context, refreshToken string) (string, error) {
	userID, err := s.client.GetDel(ctx, hashToken(refreshToken)).Result()
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

func (s *RefreshTokenStore) DeleteAllByUserID(ctx context.Context, userID string) error {
	setKey := "user_tokens:" + userID
	tokens, err := s.client.SMembers(ctx, setKey).Result()
	if err != nil {
		return err
	}
	if len(tokens) == 0 {
		return nil
	}

	pipe := s.client.TxPipeline()
	for _, t := range tokens {
		pipe.Del(ctx, t)
	}
	pipe.Del(ctx, setKey)
	_, err = pipe.Exec(ctx)
	return err
}
