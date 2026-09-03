package interceptor

import (
	"context"

	"buf.build/go/protovalidate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type ValidationInterceptor struct {
	validator protovalidate.Validator
}

func NewValidationInterceptor() (*ValidationInterceptor, error) {
	v, err := protovalidate.New()
	if err != nil {
		return nil, err
	}
	return &ValidationInterceptor{validator: v}, nil
}

func (i *ValidationInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		msg, ok := req.(protoreflect.ProtoMessage)
		if ok {
			if err := i.validator.Validate(msg); err != nil {
				return nil, status.Error(codes.InvalidArgument, err.Error())
			}
		}
		return handler(ctx, req)
	}
}
