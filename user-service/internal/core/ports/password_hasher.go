package ports

import "context"

type PasswordHasher interface {
	Hash(ctx context.Context, plainPassword string) (string, error)
	Compare(ctx context.Context, hash, plainPassword string) error
}
