package token

import (
	"context"
	"errors"

	"userservice/internal/core/domain"
	"userservice/internal/core/ports"
)

type InvalidReason string

const (
	InvalidReasonNone      InvalidReason = ""
	InvalidReasonExpired   InvalidReason = "expired"
	InvalidReasonMalformed InvalidReason = "malformed"
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
	Reason InvalidReason
}

func (uc *ValidateTokenCase) Execute(ctx context.Context, accessToken string) *ValidateTokenResult {
	userID, role, err := uc.tokens.Parse(accessToken)
	if err != nil {
		reason := InvalidReasonMalformed
		if errors.Is(err, ports.ErrTokenExpired) {
			reason = InvalidReasonExpired
		}
		return &ValidateTokenResult{Valid: false, Reason: reason}
	}
	return &ValidateTokenResult{Valid: true, UserID: userID, Role: role, Reason: InvalidReasonNone}
}
