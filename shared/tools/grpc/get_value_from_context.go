package grpc

import "context"

func ValueFromContext(ctx context.Context, key any) (string, bool) {
	value, ok := ctx.Value(key).(string)
	return value, ok
}
