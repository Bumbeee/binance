package hasher

import (
	"context"
	"market/shared/tracing"

	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/crypto/bcrypt"
)

type BCryptHasher struct {
	cost int
}

func NewBCryptHasher(cost int) *BCryptHasher {
	return &BCryptHasher{cost: cost}
}

func (h *BCryptHasher) Hash(ctx context.Context, plainPassword string) (hash string, err error) {
	_, span := tracing.Start(ctx, "hasher.Hash", attribute.Int("bcrypt.cost", h.cost))
	defer span.End(&err)

	bytes, err := bcrypt.GenerateFromPassword([]byte(plainPassword), h.cost)
	return string(bytes), err
}

func (h *BCryptHasher) Compare(ctx context.Context, hash, plainPassword string) (err error) {
	_, span := tracing.Start(ctx, "hasher.Compare")
	defer span.End(&err, bcrypt.ErrMismatchedHashAndPassword)

	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plainPassword))
}
