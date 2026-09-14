package grpc

import (
	"context"
	"errors"

	orderv1 "market/proto/orderservice/v1"
	"order-service/internal/adapters/inbound/grpc/interceptor"
	"order-service/internal/core/domain"
	"order-service/internal/core/services/order"
)

var errMissingUserID = errors.New("user id not found in context")

type Server struct {
	orderv1.UnimplementedOrderServiceServer
	createOrder *order.CreateOrderCase
	getOrder    *order.GetOrderCase
	listOrders  *order.ListOrdersCase
	cancelOrder *order.CancelOrderCase
}

func NewServer(
	createOrder *order.CreateOrderCase,
	getOrder *order.GetOrderCase,
	listOrders *order.ListOrdersCase,
	cancelOrder *order.CancelOrderCase,
) *Server {
	return &Server{
		createOrder: createOrder,
		getOrder:    getOrder,
		listOrders:  listOrders,
		cancelOrder: cancelOrder,
	}
}

func (s *Server) CreateOrder(ctx context.Context, req *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	userID, ok := interceptor.UserIDFromContext(ctx)
	if !ok {
		return nil, toGRPCError(errMissingUserID)
	}

	var price string
	if req.Price != nil {
		price = *req.Price
	}

	res, err := s.createOrder.Execute(
		ctx,
		userID,
		req.InstrumentId,
		fromProtoSide(req.Side),
		fromProtoType(req.Type),
		price,
		req.Quantity,
	)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toCreateOrderResponse(res), nil
}

func (s *Server) GetOrder(ctx context.Context, req *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	userID, ok := interceptor.UserIDFromContext(ctx)
	if !ok {
		return nil, toGRPCError(errMissingUserID)
	}

	res, err := s.getOrder.Execute(ctx, userID, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toGetOrderResponse(res), nil
}

func (s *Server) ListOrders(ctx context.Context, req *orderv1.ListOrdersRequest) (*orderv1.ListOrdersResponse, error) {
	userID, ok := interceptor.UserIDFromContext(ctx)
	if !ok {
		return nil, toGRPCError(errMissingUserID)
	}

	var statusFilter *domain.OrderStatus
	if req.StatusFilter != nil {
		s := fromProtoStatus(*req.StatusFilter)
		statusFilter = &s
	}

	res, err := s.listOrders.Execute(ctx, userID, statusFilter)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toListOrdersResponse(res), nil
}

func (s *Server) CancelOrder(ctx context.Context, req *orderv1.CancelOrderRequest) (*orderv1.CancelOrderResponse, error) {
	userID, ok := interceptor.UserIDFromContext(ctx)
	if !ok {
		return nil, toGRPCError(errMissingUserID)
	}

	res, err := s.cancelOrder.Execute(ctx, userID, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toCancelOrderResponse(res), nil
}
