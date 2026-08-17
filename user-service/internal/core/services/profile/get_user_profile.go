package profile

import (
	"context"
	"time"

	"userservice/internal/core/ports"
)

type GetUserProfileCase struct {
	repo ports.UserRepositiry
}

func NewGetUserProfileCase(repo ports.UserRepositiry) *GetUserProfileCase {
	return &GetUserProfileCase{repo: repo}
}

type GetUserProfileResult struct {
	UserID    string
	Email     string
	Role      string
	CreatedAt time.Time
}

func (uc *GetUserProfileCase) Execute(ctx context.Context, targetUserID string) (GetUserProfileResult, error) {
	u, err := uc.repo.FindByID(ctx, targetUserID)
	if err != nil {
		return GetUserProfileResult{}, err
	}
	return GetUserProfileResult{
		UserID:    u.ID.String(),
		Email:     u.Email,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt,
	}, nil
}
