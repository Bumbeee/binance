package domain

import (
	"time"
	"userservice/shared/validator"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string
	Role         Role
	CreatedAt    time.Time
}

func NewUser(rawEmail string, passwordHash string) (*User, error) {
	return &User{
		ID:           uuid.New(),
		Email:        validator.NormalizeEmail(rawEmail),
		PasswordHash: passwordHash,
		Role:         RoleUser,
		CreatedAt:    time.Now(),
	}, nil
}
