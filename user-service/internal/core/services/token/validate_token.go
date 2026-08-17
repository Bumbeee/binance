package token

import (
	"context"
	"userservice/internal/core/domain"
	"userservice/internal/core/ports"
)

type ValidateTokenCase struct {
	tokens ports.TokenIssuer
}

func NewValidateTokenCase(tokens ports.TokenIssuer) *ValidateTokenCase {
	return &ValidateTokenCase{tokens: tokens}
}

type ValidateTokenResult struct {
	Valid  bool
	UserID string
	Role   domain.Role
}

func (uc *ValidateTokenCase) Execute(ctx context.Context, accessToken string) *ValidateTokenResult {
	userID, role, err := uc.tokens.Parse(accessToken)
	if err != nil {
		return &ValidateTokenResult{Valid: false}
	}
	return &ValidateTokenResult{Valid: true, UserID: userID, Role: role}
}
