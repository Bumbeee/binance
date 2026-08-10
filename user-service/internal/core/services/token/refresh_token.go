package token

import (
	"context"
	"errors"
	"time"
	"userservice/internal/core/ports"
)

type RefreshTokenCase struct {
	store      ports.RefreshTokenStore
	token      ports.TokenIssuer
	refreshTTL time.Duration
}

func NewRefreshTokenCase(store ports.RefreshTokenStore, tokenIssuer ports.TokenIssuer, refreshTTL time.Duration) *RefreshTokenCase {
	return &RefreshTokenCase{
		store:      store,
		token:      tokenIssuer,
		refreshTTL: refreshTTL,
	}
}

type RefreshTokenResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

func (r *RefreshTokenCase) Execute(ctx context.Context, oldRefreshToken string) (*RefreshTokenResult, error) {
	userID, err := r.store.GetUserID(ctx, oldRefreshToken)
	if err != nil {
		if errors.Is(err, ports.ErrKeyNotFound) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, err
	}

	newAccessToken, expiresAt, err := r.token.Issue(userID)
	if err != nil {
		return nil, err
	}

	newRefreshToken, err := GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	if err := r.store.Save(ctx, newRefreshToken, userID, r.refreshTTL); err != nil {
		return nil, err
	}

	if err := r.store.Delete(ctx, oldRefreshToken); err != nil {
		return nil, err
	}

	return &RefreshTokenResult{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}
