package jwt

import (
	"time"
	"userservice/internal/core/domain"

	"github.com/golang-jwt/jwt/v5"
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

	t, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrWrongSighMethod
		}
		return i.secret, nil
	})
	if err != nil {
		return "", "", err
	}
	if !t.Valid {
		return "", "", ErrInvalidToken
	}
	if claims.UserID == "" {
		return "", "", ErrInvalidSubjectClaim
	}

	return claims.UserID, claims.Role, nil
}
