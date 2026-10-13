package auth

import (
	"context"
	"fmt"
	"strings"

	"market/shared/tracing"
	"market/shared/validator"
	"userservice/internal/core/ports"
)

type ChangePasswordCase struct {
	repo                 ports.UserRepository
	hasher               ports.PasswordHasher
	passwordRequirements validator.PasswordRequirements
}

func NewChangePasswordCase(
	repo ports.UserRepository,
	hasher ports.PasswordHasher,
	passwordRequirements validator.PasswordRequirements,
) *ChangePasswordCase {
	return &ChangePasswordCase{
		repo:                 repo,
		hasher:               hasher,
		passwordRequirements: passwordRequirements,
	}
}

func (uc *ChangePasswordCase) Execute(ctx context.Context, userID, oldPassword, newPassword string) (err error) {
	ctx, span := tracing.Start(ctx, "auth.ChangePassword")
	defer span.End(&err, ErrWeakPassword)

	user, err := uc.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}

	if err := uc.hasher.Compare(ctx, user.PasswordHash, oldPassword); err != nil {
		return ErrInvalidCredentials
	}

	result := validator.ValidatePassword(newPassword, uc.passwordRequirements)
	if !result.Valid {
		return fmt.Errorf("%w: %s", ErrWeakPassword, strings.Join(result.Errors, "; "))
	}

	newHash, err := uc.hasher.Hash(ctx, newPassword)
	if err != nil {
		return err
	}

	return uc.repo.UpdatePasswordHash(ctx, userID, newHash)
}
