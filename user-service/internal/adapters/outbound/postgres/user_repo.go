package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"userservice/internal/core/domain"
	"userservice/internal/core/ports"
)

type pgxRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) ports.UserRepository {
	return &pgxRepository{pool: pool}
}

func (repo *pgxRepository) Save(ctx context.Context, user *domain.User) error {
	query := `
	INSERT INTO users (id, email, password_hash, role, created_at, first_name, last_name)
	VALUES ($1, $2, $3, $4, $5, $6, $7)
`
	_, err := repo.pool.Exec(ctx, query,
		user.ID, user.Email, user.PasswordHash, string(user.Role), user.CreatedAt,
		user.FirstName, user.LastName,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrUserAlreadyExists
		}
		return fmt.Errorf("user_repo.Save: %w", err)
	}

	return nil
}

func (repo *pgxRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, role, created_at, first_name, last_name
		FROM users
		WHERE email = $1
	`

	var (
		id           uuid.UUID
		emailValue   string
		passwordHash string
		role         string
		createdAt    time.Time
		firstName    string
		lastName     string
	)

	err := repo.pool.QueryRow(ctx, query, email).
		Scan(&id, &emailValue, &passwordHash, &role, &createdAt, &firstName, &lastName)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user_repo.GetByEmail: %w", err)
	}

	return rowToUser(id, emailValue, passwordHash, role, firstName, lastName, createdAt), nil
}

func (repo *pgxRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, role, created_at, first_name, last_name
		FROM users
		WHERE id = $1
	`

	var (
		userID       uuid.UUID
		emailValue   string
		passwordHash string
		role         string
		createdAt    time.Time
		firstName    string
		lastName     string
	)

	err := repo.pool.QueryRow(ctx, query, id).
		Scan(&userID, &emailValue, &passwordHash, &role, &createdAt, &firstName, &lastName)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user_repo.FindByID: %w", err)
	}

	return rowToUser(userID, emailValue, passwordHash, role, firstName, lastName, createdAt), nil
}

func rowToUser(id uuid.UUID, email, passwordHash, role, firstName, lastName string, createdAt time.Time) *domain.User {
	return &domain.User{
		ID:           id,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         domain.Role(role),
		FirstName:    firstName,
		LastName:     lastName,
		CreatedAt:    createdAt,
	}
}

func (repo *pgxRepository) UpdatePasswordHash(ctx context.Context, userID, newPasswordHash string) error {
	query := `UPDATE users SET password_hash = $1 WHERE id = $2`

	tag, err := repo.pool.Exec(ctx, query, newPasswordHash, userID)
	if err != nil {
		return fmt.Errorf("user_repo.UpdatePasswordHash: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (repo *pgxRepository) UpdateProfile(ctx context.Context, userID string, firstName, lastName *string) (*domain.User, error) {
	query := `
		UPDATE users
		SET first_name = COALESCE($1, first_name),
		    last_name  = COALESCE($2, last_name)
		WHERE id = $3
		RETURNING id, email, password_hash, role, created_at, first_name, last_name
	`

	var (
		id           uuid.UUID
		email        string
		passwordHash string
		role         string
		createdAt    time.Time
		newFirstName string
		newLastName  string
	)

	err := repo.pool.QueryRow(ctx, query, firstName, lastName, userID).
		Scan(&id, &email, &passwordHash, &role, &createdAt, &newFirstName, &newLastName)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user_repo.UpdateProfile: %w", err)
	}

	return rowToUser(id, email, passwordHash, role, newFirstName, newLastName, createdAt), nil
}
