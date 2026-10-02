package domain

import "github.com/shopspring/decimal"

func ComputeFill(remaining, peerRemaining decimal.Decimal) decimal.Decimal {
	return decimal.Min(remaining, peerRemaining)
}

func ResolveFillStatus(remainingAfterFill decimal.Decimal) OrderStatus {
	if remainingAfterFill.IsZero() {
		return OrderStatusFilled
	}
	return OrderStatusPartiallyFilled
}

func ResolveFinalStatus(remainingAfter, originalQuantity decimal.Decimal, orderType OrderType) OrderStatus {
	switch {
	case remainingAfter.IsZero():
		return OrderStatusFilled
	case orderType == OrderTypeMarket:
		return OrderStatusCancelled
	case remainingAfter.LessThan(originalQuantity):
		return OrderStatusPartiallyFilled
	default:
		return OrderStatusOpen
	}
}
