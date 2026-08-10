package token

import (
	"context"
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
}

func (uc *ValidateTokenCase) Execute(ctx context.Context, accessToken string) *ValidateTokenResult {
	userID, err := uc.tokens.Parse(accessToken)
	if err != nil {
		return &ValidateTokenResult{Valid: false}
	}
	return &ValidateTokenResult{Valid: true, UserID: userID}
}
