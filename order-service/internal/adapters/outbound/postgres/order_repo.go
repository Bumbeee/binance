package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"order-service/internal/core/domain"
	"order-service/internal/core/ports"
)

type pgxOrderRepository struct {
	pool *pgxpool.Pool
}

func NewOrderRepository(pool *pgxpool.Pool) ports.OrderRepository {
	return &pgxOrderRepository{pool: pool}
}

func (r *pgxOrderRepository) Save(ctx context.Context, order *domain.Order) error {
	query := `
		INSERT INTO orders (
			id, user_id, instrument_id, side, type,
			price, quantity, status, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	var price any
	if order.Price != "" {
		price = order.Price
	}

	_, err := r.pool.Exec(ctx, query,
		order.ID,
		order.UserID,
		order.InstrumentID,
		string(order.Side),
		string(order.Type),
		price,
		order.Quantity,
		string(order.Status),
		order.CreatedAt,
		order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("order_repo.Save: %w", err)
	}

	return nil
}

func (r *pgxOrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	query := `
		SELECT id, user_id, instrument_id, side, type,
		       price, quantity, status, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	order, err := scanOrder(r.pool.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrOrderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("order_repo.GetByID: %w", err)
	}

	return order, nil
}

func (r *pgxOrderRepository) ListByUserID(ctx context.Context, userID uuid.UUID, statusFilter *domain.OrderStatus) ([]*domain.Order, error) {
	var (
		rows pgx.Rows
		err  error
	)

	if statusFilter != nil {
		query := `
			SELECT id, user_id, instrument_id, side, type,
			       price, quantity, status, created_at, updated_at
			FROM orders
			WHERE user_id = $1 AND status = $2
			ORDER BY created_at DESC
		`
		rows, err = r.pool.Query(ctx, query, userID, string(*statusFilter))
	} else {
		query := `
			SELECT id, user_id, instrument_id, side, type,
			       price, quantity, status, created_at, updated_at
			FROM orders
			WHERE user_id = $1
			ORDER BY created_at DESC
		`
		rows, err = r.pool.Query(ctx, query, userID)
	}
	if err != nil {
		return nil, fmt.Errorf("order_repo.ListByUserID: %w", err)
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			return nil, fmt.Errorf("order_repo.ListByUserID: %w", err)
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("order_repo.ListByUserID: %w", err)
	}

	return orders, nil
}

func (r *pgxOrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	query := `UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`

	tag, err := r.pool.Exec(ctx, query, string(status), time.Now(), id)
	if err != nil {
		return fmt.Errorf("order_repo.UpdateStatus: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrOrderNotFound
	}

	return nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanOrder(row rowScanner) (*domain.Order, error) {
	var (
		id           uuid.UUID
		userID       uuid.UUID
		instrumentID uuid.UUID
		side         string
		orderType    string
		price        *string
		quantity     string
		status       string
		createdAt    time.Time
		updatedAt    time.Time
	)

	if err := row.Scan(
		&id, &userID, &instrumentID, &side, &orderType,
		&price, &quantity, &status, &createdAt, &updatedAt,
	); err != nil {
		return nil, err
	}

	priceValue := ""
	if price != nil {
		priceValue = *price
	}

	return &domain.Order{
		ID:           id,
		UserID:       userID,
		InstrumentID: instrumentID,
		Side:         domain.OrderSide(side),
		Type:         domain.OrderType(orderType),
		Price:        priceValue,
		Quantity:     quantity,
		Status:       domain.OrderStatus(status),
		CreatedAt:    createdAt,
		UpdatedAt:    updatedAt,
	}, nil
}
