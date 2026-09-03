package ports

import (
	"context"
	"userservice/internal/core/domain"
)

type UserRepository interface {
	Save(ctx context.Context, user *domain.User) error
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByID(ctx context.Context, id string) (*domain.User, error)
	UpdatePasswordHash(ctx context.Context, userID, newPasswordHash string) error
	UpdateProfile(ctx context.Context, userID string, firstName, lastName *string) error
}
