package auth

import (
	"context"
	"market/shared/tracing"

	"userservice/internal/core/ports"
)

type LogoutCase struct {
	store ports.RefreshTokenStore
}

func NewLogoutCase(store ports.RefreshTokenStore) *LogoutCase {
	return &LogoutCase{store: store}
}

func (uc *LogoutCase) Execute(ctx context.Context, refreshToken string) (err error) {
	ctx, span := tracing.Start(ctx, "auth.Logout")
	defer span.End(&err, ErrInvalidCredentials)

	return uc.store.Delete(ctx, refreshToken)
}
