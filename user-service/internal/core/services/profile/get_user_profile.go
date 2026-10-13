package profile

import (
	"context"
	"market/shared/tracing"
	"time"

	"userservice/internal/core/domain"
	"userservice/internal/core/ports"
)

type GetUserProfileCase struct {
	repo ports.UserRepository
}

func NewGetUserProfileCase(repo ports.UserRepository) *GetUserProfileCase {
	return &GetUserProfileCase{repo: repo}
}

type GetUserProfileResult struct {
	UserID    string
	Email     string
	Role      string
	CreatedAt time.Time
}

func (uc *GetUserProfileCase) Execute(ctx context.Context, targetUserID string) (res *GetUserProfileResult, err error) {
	ctx, span := tracing.Start(ctx, "profile.GetProfile")
	defer span.End(&err, domain.ErrUserNotFound)

	u, err := uc.repo.FindByID(ctx, targetUserID)
	if err != nil {
		return nil, err
	}

	res = &GetUserProfileResult{
		UserID:    u.ID.String(),
		Email:     u.Email,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt,
	}
	return res, nil
}
