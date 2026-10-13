package profile

import (
	"context"
	"market/shared/tracing"
	"time"

	"userservice/internal/core/domain"
	"userservice/internal/core/ports"
)

type GetProfileCase struct {
	repo ports.UserRepository
}

func NewGetProfileCase(repo ports.UserRepository) *GetProfileCase {
	return &GetProfileCase{repo: repo}
}

type GetProfileResult struct {
	UserID    string
	Email     string
	Role      string
	CreatedAt time.Time
}

func (uc *GetProfileCase) Execute(ctx context.Context, userID string) (res *GetProfileResult, err error) {
	ctx, span := tracing.Start(ctx, "profile.GetProfile")
	defer span.End(&err, domain.ErrUserNotFound)

	u, err := uc.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	res = &GetProfileResult{
		UserID:    u.ID.String(),
		Email:     u.Email,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt,
	}
	return res, nil
}
