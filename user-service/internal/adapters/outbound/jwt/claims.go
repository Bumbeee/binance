package jwt

import (
	"userservice/internal/core/domain"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID string      `json:"sub"`
	Role   domain.Role `json:"role"`
	jwt.RegisteredClaims
}
