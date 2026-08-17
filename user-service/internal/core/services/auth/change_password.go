package auth

import (
	"context"
	"fmt"
	"strings"

	"userservice/internal/core/ports"
	"userservice/shared/validator"
)

type ChangePasswordCase struct {
	repo                 ports.UserRepositiry
	hasher               ports.PasswordHasher
	passwordRequirements validator.PasswordRequirements
}

func NewChangePasswordCase(
	repo ports.UserRepositiry,
	hasher ports.PasswordHasher,
	passwordRequirements validator.PasswordRequirements,
) *ChangePasswordCase {
	return &ChangePasswordCase{
		repo:                 repo,
		hasher:               hasher,
		passwordRequirements: passwordRequirements,
	}
}

func (uc *ChangePasswordCase) Execute(ctx context.Context, userID, oldPassword, newPassword string) error {
	user, err := uc.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := uc.hasher.Compare(user.PasswordHash, oldPassword); err != nil {
		return ErrInvalidCredentials
	}

	result := validator.ValidatePassword(newPassword, uc.passwordRequirements)
	if !result.Valid {
		return fmt.Errorf("%w: %s", ErrWeakPassword, strings.Join(result.Errors, "; "))
	}

	newHash, err := uc.hasher.Hash(newPassword)
	if err != nil {
		return err
	}

	return uc.repo.UpdatePasswordHash(ctx, userID, newHash)
}
