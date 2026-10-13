package token

import (
	"context"
	"errors"
	"market/shared/tracing"
	"time"

	"userservice/internal/core/ports"

	"go.opentelemetry.io/otel/attribute"
)

type RefreshTokenCase struct {
	store      ports.RefreshTokenStore
	token      ports.TokenIssuer
	repo       ports.UserRepository
	refreshTTL time.Duration
}

func NewRefreshTokenCase(
	store ports.RefreshTokenStore,
	tokenIssuer ports.TokenIssuer,
	repo ports.UserRepository,
	refreshTTL time.Duration,
) *RefreshTokenCase {
	return &RefreshTokenCase{
		store:      store,
		token:      tokenIssuer,
		repo:       repo,
		refreshTTL: refreshTTL,
	}
}

type RefreshTokenResult struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

func (r *RefreshTokenCase) Execute(ctx context.Context, oldRefreshToken string) (res *RefreshTokenResult, err error) {
	ctx, span := tracing.Start(ctx, "token.Refresh")
	defer span.End(&err, ErrInvalidRefreshToken)

	userID, err := r.store.GetAndDelete(ctx, oldRefreshToken)
	if err != nil {
		if errors.Is(err, ports.ErrKeyNotFound) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, err
	}
	span.SetAttributes(attribute.String("user.id", userID))

	user, err := r.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	newAccessToken, expiresAt, err := r.token.Issue(userID, user.Role)
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

	res = &RefreshTokenResult{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresAt:    expiresAt,
	}

	return res, nil
}
