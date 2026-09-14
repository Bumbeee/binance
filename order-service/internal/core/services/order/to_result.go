package order

import "order-service/internal/core/domain"

func toResult(o *domain.Order) *OrderResult {
	return &OrderResult{
		ID:           o.ID.String(),
		UserID:       o.UserID.String(),
		InstrumentID: o.InstrumentID.String(),
		Side:         string(o.Side),
		Type:         string(o.Type),
		Price:        o.Price,
		Quantity:     o.Quantity,
		Status:       string(o.Status),
		CreatedAt:    o.CreatedAt,
		UpdatedAt:    o.UpdatedAt,
	}
}
