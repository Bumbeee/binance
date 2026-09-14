package grpc

import (
	"google.golang.org/protobuf/types/known/timestamppb"

	orderv1 "market/proto/orderservice/v1"
	"order-service/internal/core/domain"
	"order-service/internal/core/services/order"
)

func toProtoSide(s string) orderv1.OrderSide {
	switch s {
	case "buy":
		return orderv1.OrderSide_ORDER_SIDE_BUY
	case "sell":
		return orderv1.OrderSide_ORDER_SIDE_SELL
	default:
		return orderv1.OrderSide_ORDER_SIDE_UNSPECIFIED
	}
}

func fromProtoSide(s orderv1.OrderSide) string {
	switch s {
	case orderv1.OrderSide_ORDER_SIDE_BUY:
		return string(domain.OrderSideBuy)
	case orderv1.OrderSide_ORDER_SIDE_SELL:
		return string(domain.OrderSideSell)
	default:
		return ""
	}
}

func toProtoType(t string) orderv1.OrderType {
	switch t {
	case "market":
		return orderv1.OrderType_ORDER_TYPE_MARKET
	case "limit":
		return orderv1.OrderType_ORDER_TYPE_LIMIT
	default:
		return orderv1.OrderType_ORDER_TYPE_UNSPECIFIED
	}
}

func fromProtoType(t orderv1.OrderType) string {
	switch t {
	case orderv1.OrderType_ORDER_TYPE_MARKET:
		return string(domain.OrderTypeMarket)
	case orderv1.OrderType_ORDER_TYPE_LIMIT:
		return string(domain.OrderTypeLimit)
	default:
		return ""
	}
}

func toProtoStatus(s string) orderv1.OrderStatus {
	switch s {
	case "open":
		return orderv1.OrderStatus_ORDER_STATUS_OPEN
	case "cancelled":
		return orderv1.OrderStatus_ORDER_STATUS_CANCELLED
	default:
		return orderv1.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

func fromProtoStatus(s orderv1.OrderStatus) domain.OrderStatus {
	switch s {
	case orderv1.OrderStatus_ORDER_STATUS_OPEN:
		return domain.OrderStatusOpen
	case orderv1.OrderStatus_ORDER_STATUS_CANCELLED:
		return domain.OrderStatusCancelled
	default:
		return ""
	}
}

func toProtoOrder(r *order.OrderResult) *orderv1.Order {
	return &orderv1.Order{
		Id:           r.ID,
		UserId:       r.UserID,
		InstrumentId: r.InstrumentID,
		Side:         toProtoSide(r.Side),
		Type:         toProtoType(r.Type),
		Price:        r.Price,
		Quantity:     r.Quantity,
		Status:       toProtoStatus(r.Status),
		CreatedAt:    timestamppb.New(r.CreatedAt),
		UpdatedAt:    timestamppb.New(r.UpdatedAt),
	}
}

func toCreateOrderResponse(r *order.OrderResult) *orderv1.CreateOrderResponse {
	return &orderv1.CreateOrderResponse{Order: toProtoOrder(r)}
}

func toGetOrderResponse(r *order.OrderResult) *orderv1.GetOrderResponse {
	return &orderv1.GetOrderResponse{Order: toProtoOrder(r)}
}

func toListOrdersResponse(results []*order.OrderResult) *orderv1.ListOrdersResponse {
	orders := make([]*orderv1.Order, 0, len(results))
	for _, r := range results {
		orders = append(orders, toProtoOrder(r))
	}
	return &orderv1.ListOrdersResponse{Orders: orders}
}

func toCancelOrderResponse(r *order.OrderResult) *orderv1.CancelOrderResponse {
	return &orderv1.CancelOrderResponse{Order: toProtoOrder(r)}
}
