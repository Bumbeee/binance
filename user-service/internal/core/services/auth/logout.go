package auth

import (
	"context"

	"userservice/internal/core/ports"
)

type LogoutCase struct {
	store ports.RefreshTokenStore
}

func NewLogoutCase(store ports.RefreshTokenStore) *LogoutCase {
	return &LogoutCase{store: store}
}

func (uc *LogoutCase) Execute(ctx context.Context, refreshToken string) error {
	return uc.store.Delete(ctx, refreshToken)
}
