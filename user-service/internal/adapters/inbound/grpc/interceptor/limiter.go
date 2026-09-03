package interceptor

import (
	"context"
	"math"
	"net"
	"time"

	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type Limiter struct {
	client    *redis.Client
	baseDelay time.Duration
	maxDelay  time.Duration
	failTTL   time.Duration
}

func NewLimiter(client *redis.Client, baseDelay, maxDelay, failTTL time.Duration) *Limiter {
	return &Limiter{
		client:    client,
		baseDelay: baseDelay,
		maxDelay:  maxDelay,
		failTTL:   failTTL,
	}
}

var limitedMethods = map[string]map[codes.Code]bool{
	"/userservice.v1.UserService/Login": {
		codes.Unauthenticated: true,
	},
	"/userservice.v1.UserService/Register": {
		codes.AlreadyExists: true,
	},
}

func (l *Limiter) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		triggers, tracked := limitedMethods[info.FullMethod]
		if !tracked {
			return handler(ctx, req)
		}

		ip, ipErr := extractIP(ctx)
		if ipErr != nil {
			return handler(ctx, req)
		}

		blockedKey := "limiter:blocked:" + info.FullMethod + ":" + ip
		failsKey := "limiter:fails:" + info.FullMethod + ":" + ip

		ttl, err := l.client.TTL(ctx, blockedKey).Result()
		if err == nil && ttl > 0 {
			return nil, status.Errorf(codes.ResourceExhausted, "too many attempts, retry after %s", ttl.Round(time.Second))
		}

		resp, handlerErr := handler(ctx, req)

		if handlerErr != nil {
			st, ok := status.FromError(handlerErr)
			if ok && triggers[st.Code()] {
				fails, _ := l.client.Incr(ctx, failsKey).Result()
				l.client.Expire(ctx, failsKey, l.failTTL)

				delay := l.computeDelay(fails)
				l.client.Set(ctx, blockedKey, "1", delay)
			}
			return resp, handlerErr
		}

		l.client.Del(ctx, failsKey)

		return resp, handlerErr
	}
}

func (l *Limiter) computeDelay(fails int64) time.Duration {
	delay := time.Duration(float64(l.baseDelay) * math.Pow(2, float64(fails-1)))
	if delay > l.maxDelay {
		return l.maxDelay
	}
	return delay
}

func extractIP(ctx context.Context) (string, error) {
	p, ok := peer.FromContext(ctx)
	if !ok {
		return "", status.Error(codes.Internal, "no peer info")
	}

	host, _, err := net.SplitHostPort(p.Addr.String())
	if err != nil {
		return "", err
	}
	return host, nil
}
