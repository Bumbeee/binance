package grpc

import "context"

// type ctxKey string
// const (
// userIDKey ctxKey = "userID"
// traceIDKey ctxKey = "traceID"
// )

func StringFromContext(ctx context.Context, key any) (string, bool) {
	value, ok := ctx.Value(key).(string)
	return value, ok
}
