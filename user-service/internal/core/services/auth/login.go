package auth

import (
	"context"
	"market/shared/tracing"
	"market/shared/validator"
	"time"
	"userservice/internal/core/ports"
	"userservice/internal/core/services/token"

	"go.opentelemetry.io/otel/attribute"
)

type LoginCase struct {
	repo              ports.UserRepository
	hasher            ports.PasswordHasher
	tokens            ports.TokenIssuer
	refreshTokenStore ports.RefreshTokenStore
	refreshTTL        time.Duration
	dummyHash         string
}

func NewLoginCase(
	repo ports.UserRepository,
	hasher ports.PasswordHasher,
	tokens ports.TokenIssuer,
	refreshTokenStore ports.RefreshTokenStore,
	refreshTTL time.Duration,
) (*LoginCase, error) {
	dummyHash, err := hasher.Hash(context.Background(), "dummy-password-never-used-for-timing-safety")
	if err != nil {
		return nil, err
	}

	return &LoginCase{
		repo:              repo,
		hasher:            hasher,
		tokens:            tokens,
		refreshTokenStore: refreshTokenStore,
		refreshTTL:        refreshTTL,
		dummyHash:         dummyHash,
	}, nil
}

type LoginResult struct {
	UserID       string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

func (lc *LoginCase) Execute(ctx context.Context, email, plainPassword string) (res *LoginResult, err error) {
	ctx, span := tracing.Start(ctx, "auth.Login")
	defer span.End(&err, ErrInvalidCredentials)

	if !validator.IsValidEmail(email) || plainPassword == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := lc.repo.GetByEmail(ctx, email)
	if err != nil {
		_ = lc.hasher.Compare(ctx, lc.dummyHash, plainPassword)
		return nil, ErrInvalidCredentials
	}

	if err := lc.hasher.Compare(ctx, user.PasswordHash, plainPassword); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, expiresAt, err := lc.tokens.Issue(user.ID.String(), user.Role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := token.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	if err := lc.refreshTokenStore.Save(ctx, refreshToken, user.ID.String(), lc.refreshTTL); err != nil {
		return nil, err
	}

	res = &LoginResult{
		UserID:       user.ID.String(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}
	span.SetAttributes(attribute.String("user.id", res.UserID))

	return res, nil
}
