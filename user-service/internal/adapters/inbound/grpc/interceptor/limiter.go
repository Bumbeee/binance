package interceptor

import (
	"context"
	"math"
	"net"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type Limiter struct {
	client    *redis.Client
	log       *zap.Logger
	baseDelay time.Duration
	maxDelay  time.Duration
	failTTL   time.Duration
}

func NewLimiter(client *redis.Client, log *zap.Logger, baseDelay, maxDelay, failTTL time.Duration) *Limiter {
	return &Limiter{
		client:    client,
		log:       log,
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
		if err != nil {
			l.log.Warn("rate limiter: failed to check block status, allowing request",
				zap.String("method", info.FullMethod), zap.Error(err))
		} else if ttl > 0 {
			return nil, status.Errorf(codes.ResourceExhausted, "too many attempts, retry after %s", ttl.Round(time.Second))
		}

		resp, handlerErr := handler(ctx, req)

		if handlerErr != nil {
			st, ok := status.FromError(handlerErr)
			if ok && triggers[st.Code()] {
				fails, err := l.client.Incr(ctx, failsKey).Result()
				if err != nil {
					l.log.Warn("rate limiter: failed to increment failure counter",
						zap.String("method", info.FullMethod), zap.Error(err))
					return resp, handlerErr
				}

				if err := l.client.Expire(ctx, failsKey, l.failTTL).Err(); err != nil {
					l.log.Warn("rate limiter: failed to set failure counter TTL",
						zap.String("method", info.FullMethod), zap.Error(err))
				}

				delay := l.computeDelay(fails)
				if err := l.client.Set(ctx, blockedKey, "1", delay).Err(); err != nil {
					l.log.Warn("rate limiter: failed to set block, backoff not applied",
						zap.String("method", info.FullMethod), zap.Error(err))
				}
			}
			return resp, handlerErr
		}

		if err := l.client.Del(ctx, failsKey).Err(); err != nil {
			l.log.Warn("rate limiter: failed to reset failure counter after success",
				zap.String("method", info.FullMethod), zap.Error(err))
		}

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
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if forwarded := md.Get("x-forwarded-for"); len(forwarded) > 0 {
			ips := strings.Split(forwarded[0], ",")
			if ip := strings.TrimSpace(ips[0]); ip != "" {
				return ip, nil
			}
		}
	}

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
