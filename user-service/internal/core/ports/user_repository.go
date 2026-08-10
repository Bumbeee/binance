package ports

import (
	"context"
	"userservice/internal/core/domain"
)

type UserRepositiry interface {
	Save(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
}
