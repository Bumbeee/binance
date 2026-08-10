package jwt

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Issuer struct {
	secret []byte
	ttl    time.Duration
}

func NewIssuer(secret string, ttl time.Duration) *Issuer {
	return &Issuer{secret: []byte(secret), ttl: ttl}
}

func (i *Issuer) Issue(userID string) (token string, expiresAt time.Time, err error) {
	expiresAt = time.Now().Add(i.ttl)

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": expiresAt.Unix(),
	})

	signed, err := t.SignedString(i.secret)
	if err != nil {
		return "", time.Time{}, err
	}
	return signed, expiresAt, nil
}

func (i *Issuer) Parse(token string) (userID string, err error) {
	t, err := jwt.Parse(token, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrWrongSighMethod
		}
		return i.secret, nil
	})
	if err != nil {
		return "", err
	}
	claims, ok := t.Claims.(jwt.MapClaims)
	if !ok || !t.Valid {
		return "", ErrInvalidToken
	}
	sub, ok := claims["sub"].(string)
	if !ok {
		return "", ErrInvalidSubjectClaim
	}
	return sub, nil
}
