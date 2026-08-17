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

func NewUserRepository(pool *pgxpool.Pool) ports.UserRepositiry {
	return &pgxRepository{pool: pool}
}

func (repo *pgxRepository) Save(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, role, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := repo.pool.Exec(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		string(user.Role),
		user.CreatedAt,
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
		SELECT id, email, password_hash, role, created_at
		FROM users
		WHERE email = $1
	`

	var (
		id           string
		emailValue   string
		passwordHash string
		role         string
		createdAt    time.Time
	)

	err := repo.pool.QueryRow(ctx, query, email).
		Scan(&id, &emailValue, &passwordHash, &role, &createdAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user_repo.GetByEmail: %w", err)
	}

	return rowToUser(id, emailValue, passwordHash, role, createdAt)
}

func (repo *pgxRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, role, created_at
		FROM users
		WHERE id = $1
	`

	var (
		userID       string
		emailValue   string
		passwordHash string
		role         string
		createdAt    time.Time
	)

	err := repo.pool.QueryRow(ctx, query, id).
		Scan(&userID, &emailValue, &passwordHash, &role, &createdAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("user_repo.FindByID: %w", err)
	}

	return rowToUser(userID, emailValue, passwordHash, role, createdAt)
}

func rowToUser(id, email, passwordHash, role string, createdAt time.Time) (*domain.User, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("user_repo: invalid user id in db: %w", err)
	}

	return &domain.User{
		ID:           parsedID,
		Email:        email,
		PasswordHash: passwordHash,
		Role:         domain.Role(role),
		CreatedAt:    createdAt,
	}, nil
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
