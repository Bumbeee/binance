package orderclient

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"

	orderv1 "market/proto/orderservice/v1"
	"spot-instrument-service/internal/core/ports"
)

type Client struct {
	stub    orderv1.OrderServiceClient
	conn    *grpc.ClientConn
	timeout time.Duration
}

func New(addr string, timeout time.Duration) (*Client, error) {
	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		return nil, err
	}

	conn.Connect()

	warmupCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	for {
		state := conn.GetState()
		if state == connectivity.Ready {
			break
		}
		if !conn.WaitForStateChange(warmupCtx, state) {
			conn.Close()
			return nil, fmt.Errorf("userclient: connection to %s did not become ready within warmup period (last state: %s)", addr, state)
		}
	}

	return &Client{
		stub:    orderv1.NewOrderServiceClient(conn),
		conn:    conn,
		timeout: timeout,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) GetLastTradePrices(ctx context.Context, instrumentIDs []string) ([]ports.PriceUpdate, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.stub.GetLastTradePrices(ctx, &orderv1.GetLastTradePricesRequest{
		InstrumentIds: instrumentIDs,
	})
	if err != nil {
		return nil, err
	}

	updates := make([]ports.PriceUpdate, 0, len(resp.Prices))
	for _, p := range resp.Prices {
		updates = append(updates, ports.PriceUpdate{
			InstrumentID: p.InstrumentId,
			Price:        p.Price,
			TradedAt:     p.TradedAt.AsTime(),
		})
	}

	return updates, nil
}
