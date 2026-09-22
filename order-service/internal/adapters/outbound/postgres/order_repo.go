package postgres

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"order-service/internal/core/domain"
	"order-service/internal/core/ports"
)

type pgxOrderRepository struct {
	pool *pgxpool.Pool
}

const defaultPageSize = 20 // в конфиг???

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
		order.ID, order.UserID, order.InstrumentID,
		string(order.Side), string(order.Type),
		price, order.Quantity, string(order.Status),
		order.CreatedAt, order.UpdatedAt,
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

func (r *pgxOrderRepository) ListByUserID(
	ctx context.Context,
	userID uuid.UUID,
	statusFilter *domain.OrderStatus,
	pageSize int32,
	pageToken string,
) ([]*domain.Order, string, error) {

	if pageSize <= 0 {
		pageSize = defaultPageSize
	}

	var (
		cursorCreatedAt time.Time
		cursorID        uuid.UUID
		hasCursor       bool
	)

	if pageToken != "" {
		var err error
		cursorCreatedAt, cursorID, err = decodeCursor(pageToken)
		if err != nil {
			return nil, "", fmt.Errorf("order_repo.ListByUserID: %w", err)
		}
		hasCursor = true
	}

	args := []any{userID}
	conditions := []string{"user_id = $1"}

	if statusFilter != nil {
		args = append(args, string(*statusFilter))
		conditions = append(conditions, fmt.Sprintf("status = $%d", len(args)))
	}
	if hasCursor {
		args = append(args, cursorCreatedAt, cursorID)
		conditions = append(conditions, fmt.Sprintf("(created_at, id) < ($%d, $%d)", len(args)-1, len(args)))
	}

	args = append(args, pageSize+1)
	query := fmt.Sprintf(`
		SELECT id, user_id, instrument_id, side, type,
		       price, quantity, status, created_at, updated_at
		FROM orders
		WHERE %s
		ORDER BY created_at DESC, id DESC
		LIMIT $%d
	`, strings.Join(conditions, " AND "), len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, "", fmt.Errorf("order_repo.ListByUserID: %w", err)
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		order, err := scanOrder(rows)
		if err != nil {
			return nil, "", fmt.Errorf("order_repo.ListByUserID: %w", err)
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("order_repo.ListByUserID: %w", err)
	}

	nextPageToken := ""
	if int32(len(orders)) > pageSize {
		last := orders[pageSize-1]
		nextPageToken = encodeCursor(last.CreatedAt, last.ID)
		orders = orders[:pageSize]
	}

	return orders, nextPageToken, nil
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

// TODO: Add List Trades(as it was added for spot-trading) to display trades for user (sillimar to ListByUserID)

func (r *pgxOrderRepository) SaveAndMatch(ctx context.Context, order *domain.Order) (*ports.MatchResult, error) {
	var result *ports.MatchResult

	err := pgx.BeginTxFunc(ctx, r.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		insertQuery := `
			INSERT INTO orders (
				id, user_id, instrument_id, side, type,
				price, quantity, remaining_quantity, status, created_at, updated_at
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		`
		var price any
		if order.Price != "" {
			price = order.Price
		}

		if _, err := tx.Exec(ctx, insertQuery,
			order.ID, order.UserID, order.InstrumentID,
			string(order.Side), string(order.Type),
			price, order.Quantity, order.RemainingQuantity, string(order.Status),
			order.CreatedAt, order.UpdatedAt,
		); err != nil {
			return fmt.Errorf("insert order: %w", err)
		}

		remaining, err := decimal.NewFromString(order.RemainingQuantity)
		if err != nil {
			return fmt.Errorf("parse order quantity: %w", err)
		}

		var (
			trades  []*domain.Trade
			peers   []*domain.Order
			zero    = decimal.NewFromInt(0)
			orderPx decimal.Decimal
		)

		if order.Type == domain.OrderTypeLimit {
			orderPx, err = decimal.NewFromString(order.Price)
			if err != nil {
				return fmt.Errorf("parse order price: %w", err)
			}
		}

		for remaining.GreaterThan(zero) {
			peer, err := lockBestMatchingOrder(ctx, tx, order.InstrumentID, order.Side, order.Type, orderPx)
			if err != nil {
				return fmt.Errorf("lock matching order: %w", err)
			}
			if peer == nil {
				break
			}

			peerRemaining, err := decimal.NewFromString(peer.RemainingQuantity)
			if err != nil {
				return fmt.Errorf("parse peer quantity: %w", err)
			}

			fillQty := decimal.Min(remaining, peerRemaining)

			trade := &domain.Trade{
				ID:           uuid.New(),
				InstrumentID: order.InstrumentID,
				Price:        peer.Price,
				Quantity:     fillQty.String(),
				CreatedAt:    time.Now(),
			}
			if order.Side == domain.OrderSideBuy {
				trade.BuyOrderID = order.ID
				trade.SellOrderID = peer.ID
			} else {
				trade.BuyOrderID = peer.ID
				trade.SellOrderID = order.ID
			}

			if _, err := tx.Exec(ctx, `
				INSERT INTO trades (id, instrument_id, buy_order_id, sell_order_id, price, quantity, created_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7)
			`, trade.ID, trade.InstrumentID, trade.BuyOrderID, trade.SellOrderID, trade.Price, trade.Quantity, trade.CreatedAt); err != nil {
				return fmt.Errorf("insert trade: %w", err)
			}
			trades = append(trades, trade)

			// TODO(wallet): once WalletService is implemented, this is
			// where each side of the trade settles:
			//   - SettleReservation(peer's reservation, fillQty) — debits
			//     what the maker actually spent from their locked funds
			//   - Credit(order.UserID, <asset received>, fillQty) and
			//     Credit(peer.UserID, <asset received>, fillQty * trade.Price)
			//     — the asset each side receives depends on which side
			//     (buy/sell) it was and needs the instrument's
			//     base/quote assets, which this repository doesn't know —
			//     that lookup and the wallet calls belong in the
			//     application layer once wired, not here.

			remaining = remaining.Sub(fillQty)
			peerRemaining = peerRemaining.Sub(fillQty)

			peerStatus := domain.OrderStatusPartiallyFilled
			if peerRemaining.IsZero() {
				peerStatus = domain.OrderStatusFilled
			}
			if _, err := tx.Exec(ctx, `
				UPDATE orders SET remaining_quantity = $1, status = $2, updated_at = $3 WHERE id = $4
			`, peerRemaining.String(), string(peerStatus), time.Now(), peer.ID); err != nil {
				return fmt.Errorf("update peer order: %w", err)
			}
			peer.RemainingQuantity = peerRemaining.String()
			peer.Status = peerStatus
			peers = append(peers, peer)
		}

		order.RemainingQuantity = remaining.String()
		switch {
		case remaining.IsZero():
			order.Status = domain.OrderStatusFilled
		case order.Type == domain.OrderTypeMarket:
			order.Status = domain.OrderStatusCancelled
		case remaining.LessThan(decimal.RequireFromString(order.Quantity)):
			order.Status = domain.OrderStatusPartiallyFilled
		default:
			order.Status = domain.OrderStatusOpen
		}
		order.UpdatedAt = time.Now()

		if _, err := tx.Exec(ctx, `
			UPDATE orders SET remaining_quantity = $1, status = $2, updated_at = $3 WHERE id = $4
		`, order.RemainingQuantity, string(order.Status), order.UpdatedAt, order.ID); err != nil {
			return fmt.Errorf("update new order: %w", err)
		}

		result = &ports.MatchResult{Order: order, Trades: trades, AffectedPeers: peers}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("order_repo.SaveAndMatch: %w", err)
	}

	return result, nil
}

func lockBestMatchingOrder(
	ctx context.Context,
	tx pgx.Tx,
	instrumentID uuid.UUID,
	side domain.OrderSide,
	orderType domain.OrderType,
	limitPrice decimal.Decimal,
) (*domain.Order, error) {
	opposite := side.Opposite()

	priceOrder := "ASC"
	priceFilter := ""
	args := []any{instrumentID, string(opposite)}

	if opposite == domain.OrderSideBuy {
		priceOrder = "DESC"
	}
	if orderType == domain.OrderTypeLimit {
		args = append(args, limitPrice.String())
		if side == domain.OrderSideBuy {
			priceFilter = fmt.Sprintf("AND price <= $%d", len(args))
		} else {
			priceFilter = fmt.Sprintf("AND price >= $%d", len(args))
		}
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, instrument_id, side, type,
		       price, quantity, remaining_quantity, status, created_at, updated_at
		FROM orders
		WHERE instrument_id = $1
		  AND side = $2
		  AND status IN ('open', 'partially_filled')
		  %s
		ORDER BY price %s, created_at ASC
		LIMIT 1
		FOR UPDATE SKIP LOCKED
	`, priceFilter, priceOrder)

	order, err := scanOrder(tx.QueryRow(ctx, query, args...))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return order, nil
}

func encodeCursor(createdAt time.Time, id uuid.UUID) string {
	raw := createdAt.Format(time.RFC3339Nano) + "|" + id.String()
	return base64.URLEncoding.EncodeToString([]byte(raw))
}

func decodeCursor(token string) (time.Time, uuid.UUID, error) {
	raw, err := base64.URLEncoding.DecodeString(token)
	if err != nil {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid page token: %w", err)
	}

	parts := strings.SplitN(string(raw), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, uuid.Nil, errors.New("invalid page token: malformed")
	}

	createdAt, err := time.Parse(time.RFC3339Nano, parts[0])
	if err != nil {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid page token: bad timestamp: %w", err)
	}

	id, err := uuid.Parse(parts[1])
	if err != nil {
		return time.Time{}, uuid.Nil, fmt.Errorf("invalid page token: bad id: %w", err)
	}

	return createdAt, id, nil
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
