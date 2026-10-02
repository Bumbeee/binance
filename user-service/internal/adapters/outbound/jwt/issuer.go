package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"userservice/internal/core/domain"
	"userservice/internal/core/ports"
)

type Issuer struct {
	secret []byte
	ttl    time.Duration
}

func NewIssuer(secret string, ttl time.Duration) *Issuer {
	return &Issuer{secret: []byte(secret), ttl: ttl}
}

func (i *Issuer) Issue(userID string, role domain.Role) (token string, expiresAt time.Time, err error) {
	expiresAt = time.Now().Add(i.ttl)

	claims := Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := t.SignedString(i.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

func (i *Issuer) Parse(tokenString string) (userID string, role domain.Role, err error) {
	claims := &Claims{}

	t, parseErr := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrWrongSignMethod
		}
		return i.secret, nil
	})

	if parseErr != nil {
		if errors.Is(parseErr, jwt.ErrTokenExpired) {
			return "", "", ports.ErrTokenExpired
		}
		return "", "", ports.ErrMalformedToken
	}
	if !t.Valid {
		return "", "", ports.ErrMalformedToken
	}
	if claims.UserID == "" {
		return "", "", ports.ErrMalformedToken
	}

	return claims.UserID, claims.Role, nil
}
