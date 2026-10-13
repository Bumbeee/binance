package order

import (
	"context"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"order-service/internal/core/domain"
	"order-service/internal/core/ports"
)

type CancelOrderCase struct {
	repo ports.OrderRepository
}

func NewCancelOrderCase(repo ports.OrderRepository) *CancelOrderCase {
	return &CancelOrderCase{repo: repo}
}

func (uc *CancelOrderCase) Execute(ctx context.Context, callerUserID, orderID string) (*OrderResult, error) {
	parsedCallerID, err := uuid.Parse(callerUserID)
	if err != nil {
		return nil, err
	}

	parsedOrderID, err := uuid.Parse(orderID)
	if err != nil {
		return nil, err
	}

	span := trace.SpanFromContext(ctx)
	span.SetAttributes(attribute.String("order.id", parsedOrderID.String()))

	o, err := uc.repo.GetByID(ctx, parsedOrderID)
	if err != nil {
		return nil, err
	}
	span.SetAttributes(attribute.String("order.status_before", string(o.Status)))

	if o.UserID != parsedCallerID {
		span.AddEvent("order.access_denied")
		return nil, domain.ErrNotOrderOwner
	}

	if err := o.Cancel(); err != nil {
		return nil, err
	}

	if err := uc.repo.UpdateStatus(ctx, parsedOrderID, o.Status); err != nil {
		return nil, err
	}

	return toResult(o), nil
}
