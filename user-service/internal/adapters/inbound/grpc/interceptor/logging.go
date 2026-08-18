package interceptor

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type LoggingInterceptor struct {
	log *zap.Logger
}

func NewLoggingInterceptor(log *zap.Logger) *LoggingInterceptor {
	return &LoggingInterceptor{log: log}
}

func (l *LoggingInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)

		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.Int64("duration_ms", duration.Milliseconds()),
		}

		if err == nil {
			l.log.Info("grpc request completed", append(fields, zap.String("code", codes.OK.String()))...)
			return resp, err
		}

		stat, ok := status.FromError(err)
		code := codes.Unknown
		if ok {
			code = stat.Code()
		}
		fields = append(fields, zap.String("code", code.String()), zap.String("error", stat.Message()))

		switch code {
		case codes.Internal, codes.Unknown, codes.DataLoss, codes.Unavailable:
			l.log.Error("grpc request failed", fields...)
		case codes.Unauthenticated, codes.PermissionDenied, codes.InvalidArgument, codes.NotFound, codes.AlreadyExists:
			l.log.Warn("grpc request rejected", fields...)
		default:
			l.log.Info("grpc request completed with error", fields...)
		}

		return resp, err
	}
}
