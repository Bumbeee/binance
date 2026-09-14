package ports

import (
	"context"
)

type TokenValidator interface {
	Validate(ctx context.Context, accessToken string) (userID string, role string, valid bool, err error)
}
