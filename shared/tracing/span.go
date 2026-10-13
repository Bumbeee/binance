package tracing

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

const tracerName = "market/usecase"

type Span struct {
	trace.Span
}

func Start(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, *Span) {
	ctx, s := otel.Tracer(tracerName).Start(ctx, name, trace.WithAttributes(attrs...))
	return ctx, &Span{Span: s}
}

func (s *Span) End(err *error, expected ...error) {
	if err != nil && *err != nil {
		s.RecordError(*err)
		if !isExpected(*err, expected) {
			s.SetStatus(codes.Error, (*err).Error())
		}
	}
	s.Span.End()
}

func isExpected(err error, expected []error) bool {
	for _, e := range expected {
		if errors.Is(err, e) {
			return true
		}
	}
	return false
}
