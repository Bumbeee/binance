package profile

import (
	"context"
	"time"
	"userservice/internal/core/domain"
	"userservice/internal/core/ports"
)

type UpdateProfileCase struct {
	repo ports.UserRepository
}

func NewUpdateProfileCase(repo ports.UserRepository) *UpdateProfileCase {
	return &UpdateProfileCase{repo: repo}
}

type UpdateProfileResult struct {
	UserID    string
	Email     string
	Role      string
	FirstName string
	LastName  string
	CreatedAt time.Time
}

func (up *UpdateProfileCase) Execute(ctx context.Context, userID string, firstName, lastName *string) (UpdateProfileResult, error) {
	if firstName != nil && !domain.ValidateName(*firstName) {
		return UpdateProfileResult{}, domain.ErrEmptyFirstName
	}
	if lastName != nil && !domain.ValidateName(*lastName) {
		return UpdateProfileResult{}, domain.ErrEmptyLastName
	}

	if err := up.repo.UpdateProfile(ctx, userID, firstName, lastName); err != nil {
		return UpdateProfileResult{}, err
	}

	u, err := up.repo.FindByID(ctx, userID)
	if err != nil {
		return UpdateProfileResult{}, err
	}

	return UpdateProfileResult{
		UserID:    u.ID.String(),
		Email:     u.Email,
		Role:      string(u.Role),
		FirstName: u.FirstName,
		LastName:  u.LastName,
		CreatedAt: u.CreatedAt,
	}, nil
}
