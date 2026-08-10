package profile

import (
	"context"
	"time"

	"userservice/internal/core/ports"
)

type GetProfileCase struct {
	repo ports.UserRepositiry
}

func NewGetProfileCase(repo ports.UserRepositiry) *GetProfileCase {
	return &GetProfileCase{repo: repo}
}

type GetProfileResult struct {
	UserID    string
	Email     string
	Role      string
	CreatedAt time.Time
}

func (uc *GetProfileCase) Execute(ctx context.Context, userID string) (*GetProfileResult, error) {
	u, err := uc.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return &GetProfileResult{
		UserID:    u.ID.String(),
		Email:     u.Email,
		Role:      string(u.Role),
		CreatedAt: u.CreatedAt,
	}, nil
}
