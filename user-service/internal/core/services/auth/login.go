package auth

import (
	"context"
	"log"
	"time"
	"userservice/internal/core/ports"
	"userservice/internal/core/services/token"
	"userservice/shared/validator"
)

type LoginCase struct {
	repo              ports.UserRepositiry
	hasher            ports.PasswordHasher
	tokens            ports.TokenIssuer
	refreshTokenStore ports.RefreshTokenStore
	refreshTTL        time.Duration
	dummyHash         string
}

func NewLoginCase(
	repo ports.UserRepositiry,
	hasher ports.PasswordHasher,
	tokens ports.TokenIssuer,
	refreshTokenStore ports.RefreshTokenStore,
	refreshTTL time.Duration,
) (*LoginCase, error) {
	dummyHash, err := hasher.Hash("dummy-password-never-used-for-timing-safety")
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

func (lc *LoginCase) Execute(ctx context.Context, email, plainPassword string) (*LoginResult, error) {
	log.Println("EXECUTE ENTERED")
	if !validator.IsValidEmail(email) || plainPassword == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := lc.repo.GetByEmail(ctx, email)
	if err != nil {
		_ = lc.hasher.Compare(lc.dummyHash, plainPassword)
		return nil, ErrInvalidCredentials
	}

	if err := lc.hasher.Compare(user.PasswordHash, plainPassword); err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, expiresAt, err := lc.tokens.Issue(user.ID.String())
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

	return &LoginResult{
		UserID:       user.ID.String(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}
