package redis

import "github.com/redis/go-redis/v9"

func NewRedisClient(redisAddr string) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
}
