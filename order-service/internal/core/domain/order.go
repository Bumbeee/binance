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
	OrderStatusOpen      OrderStatus = "open"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	InstrumentID uuid.UUID
	Side         OrderSide
	Type         OrderType
	Price        string
	Quantity     string
	Status       OrderStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
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

	return &Order{
		ID:           uuid.New(),
		UserID:       userID,
		InstrumentID: instrumentID,
		Side:         side,
		Type:         orderType,
		Price:        price,
		Quantity:     quantity,
		Status:       OrderStatusOpen,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

func (o *Order) Cancel() error {
	if o.Status == OrderStatusCancelled {
		return ErrOrderAlreadyCancelled
	}
	o.Status = OrderStatusCancelled
	o.UpdatedAt = time.Now()
	return nil
}
