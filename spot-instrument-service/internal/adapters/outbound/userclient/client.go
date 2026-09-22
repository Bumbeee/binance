package userclient

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	user "market/proto/userservice/v1"
)

type Client struct {
	stub    user.UserServiceClient
	conn    *grpc.ClientConn
	timeout time.Duration
}

func New(addr string, timeout time.Duration) (*Client, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &Client{
		stub:    user.NewUserServiceClient(conn),
		conn:    conn,
		timeout: timeout,
	}, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}

func (c *Client) Validate(ctx context.Context, accessToken string) (userID string, role string, valid bool, err error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	resp, err := c.stub.ValidateToken(ctx, &user.ValidateTokenRequest{
		AccessToken: accessToken,
	})

	if err != nil {
		return "", "", false, err
	}

	if !resp.Valid {
		return "", "", false, nil
	}

	return resp.UserId, resp.Role.String(), true, nil
}
