package ports

import "time"

type TokenIssuer interface {
	Issue(userID string) (token string, expiresAt time.Time, err error)
	Parse(token string) (userID string, err error)
}
