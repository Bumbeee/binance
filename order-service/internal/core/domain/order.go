package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrderSide string

const (
	OrderSideBuy  OrderSide = "buy"
	OrderSideSell OrderSide = "sell"
)

func (s OrderSide) IsValid() bool {
	return s == OrderSideBuy || s == OrderSideSell
}

func (s OrderSide) Opposite() OrderSide {
	if s == OrderSideBuy {
		return OrderSideSell
	}
	return OrderSideBuy
}

type OrderType string

const (
	OrderTypeMarket OrderType = "market"
	OrderTypeLimit  OrderType = "limit"
)

func (t OrderType) IsValid() bool {
	return t == OrderTypeMarket || t == OrderTypeLimit
}

type OrderStatus string

const (
	OrderStatusOpen            OrderStatus = "open"
	OrderStatusPartiallyFilled OrderStatus = "partially_filled"
	OrderStatusFilled          OrderStatus = "filled"
	OrderStatusCancelled       OrderStatus = "cancelled"
)

func (s OrderStatus) IsTerminal() bool {
	return s == OrderStatusFilled || s == OrderStatusCancelled
}

type Order struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	InstrumentID      uuid.UUID
	Side              OrderSide
	Type              OrderType
	Price             string // empty for market orders
	Quantity          string
	RemainingQuantity string
	Status            OrderStatus
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

func NewOrder(userID, instrumentID uuid.UUID, side OrderSide, orderType OrderType, price, quantity string) (*Order, error) {
	if !side.IsValid() {
		return nil, ErrInvalidSide
	}
	if !orderType.IsValid() {
		return nil, ErrInvalidType
	}
	if quantity == "" {
		return nil, ErrEmptyQuantity
	}

	switch orderType {
	case OrderTypeLimit:
		if price == "" {
			return nil, ErrPriceRequiredForLimit
		}
	case OrderTypeMarket:
		if price != "" {
			return nil, ErrPriceNotAllowedForMarket
		}
	}

	now := time.Now()

	return &Order{
		ID:                uuid.New(),
		UserID:            userID,
		InstrumentID:      instrumentID,
		Side:              side,
		Type:              orderType,
		Price:             price,
		Quantity:          quantity,
		RemainingQuantity: quantity,
		Status:            OrderStatusOpen,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

func (o *Order) Cancel() error {
	if o.Status.IsTerminal() {
		return ErrOrderAlreadyTerminal
	}
	o.Status = OrderStatusCancelled
	o.UpdatedAt = time.Now()
	return nil
}
