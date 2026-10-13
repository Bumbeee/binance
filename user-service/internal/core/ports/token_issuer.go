package ports

import (
	"context"
	"time"
	"userservice/internal/core/domain"
)

type TokenIssuer interface {
	Issue(userID string, role domain.Role) (token string, expiresAt time.Time, err error)
	Parse(ctx context.Context, tokenString string) (userID string, role domain.Role, err error)
}
