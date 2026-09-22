package order

import "time"

type OrderResult struct {
	ID                string
	UserID            string
	InstrumentID      string
	Side              string
	Type              string
	Price             string
	Quantity          string
	RemainingQuantity string
	Status            string
	CreatedAt         time.Time
	UpdatedAt         time.Time
}
