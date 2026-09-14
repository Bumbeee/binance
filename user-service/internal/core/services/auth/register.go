package auth

import (
	"context"
	"market/shared/validator"
	"time"
	"userservice/internal/core/domain"
	"userservice/internal/core/ports"
	"userservice/internal/core/services/token"
)

type RegisterCase struct {
	repo                 ports.UserRepository
	hasher               ports.PasswordHasher
	tokens               ports.TokenIssuer
	refreshTokenStore    ports.RefreshTokenStore
	refreshTTL           time.Duration
	passwordRequirements validator.PasswordRequirements
}

func NewRegisterCase(
	repo ports.UserRepository,
	hasher ports.PasswordHasher,
	tokens ports.TokenIssuer,
	refreshTokenStore ports.RefreshTokenStore,
	refreshTTL time.Duration,
	passwordRequirements validator.PasswordRequirements) *RegisterCase {
	return &RegisterCase{
		repo:                 repo,
		hasher:               hasher,
		tokens:               tokens,
		refreshTokenStore:    refreshTokenStore,
		refreshTTL:           refreshTTL,
		passwordRequirements: passwordRequirements}
}

type RegisterResult struct {
	UserID       string
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

func (r *RegisterCase) Execute(ctx context.Context, email, plainPassword, firstName, lastName string) (*RegisterResult, error) {
	if !ValidateCreds(email, plainPassword, r.passwordRequirements) {
		return nil, domain.ErrInvalidCreds
	}

	hash, err := r.hasher.Hash(plainPassword)
	if err != nil {
		return nil, err
	}

	user, err := domain.NewUser(email, hash, firstName, lastName)
	if err != nil {
		return nil, err
	}

	if err := r.repo.Save(ctx, user); err != nil {
		return nil, err
	}

	accessToken, expiresAt, err := r.tokens.Issue(user.ID.String(), user.Role)
	if err != nil {
		return nil, err
	}

	refreshToken, err := token.GenerateRefreshToken()
	if err != nil {
		return nil, err
	}

	if err := r.refreshTokenStore.Save(ctx, refreshToken, user.ID.String(), r.refreshTTL); err != nil {
		return nil, err
	}

	return &RegisterResult{
		UserID:       user.ID.String(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}
