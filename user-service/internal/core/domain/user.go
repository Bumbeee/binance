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
	FirstName    string
	LastName     string
	Role         Role
	CreatedAt    time.Time
}

func NewUser(rawEmail, passwordHash, firstName, lastName string) (*User, error) {
	if !ValidateName(firstName) {
		return nil, ErrEmptyFirstName
	}
	if !ValidateName(lastName) {
		return nil, ErrEmptyLastName
	}

	return &User{
		ID:           uuid.New(),
		Email:        validator.NormalizeEmail(rawEmail),
		PasswordHash: passwordHash,
		FirstName:    firstName,
		LastName:     lastName,
		Role:         RoleUser,
		CreatedAt:    time.Now(),
	}, nil
}

func ValidateName(name string) bool {
	return name != ""
}
