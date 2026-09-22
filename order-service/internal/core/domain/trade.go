package domain

import (
	"time"

	"github.com/google/uuid"
)

type Trade struct {
	ID           uuid.UUID
	InstrumentID uuid.UUID
	BuyOrderID   uuid.UUID
	SellOrderID  uuid.UUID
	Price        string
	Quantity     string
	CreatedAt    time.Time
}
