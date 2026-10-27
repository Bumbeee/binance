package kafka

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type Record struct {
	Topic     string
	Partition int32
	Offset    int64
	Key       []byte
	Value     []byte
	Timestamp time.Time
	Headers   map[string]string
}

func toRecord(m *sarama.ConsumerMessage) *Record {
	h := make(map[string]string, len(m.Headers))
	for _, kv := range m.Headers {
		h[string(kv.Key)] = string(kv.Value)
	}
	return &Record{
		Topic: m.Topic, Partition: m.Partition, Offset: m.Offset,
		Key: m.Key, Value: m.Value, Timestamp: m.Timestamp, Headers: h,
	}
}

func injectHeaders(ctx context.Context) []sarama.RecordHeader {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	out := make([]sarama.RecordHeader, 0, len(carrier))
	for k, v := range carrier {
		out = append(out, sarama.RecordHeader{Key: []byte(k), Value: []byte(v)})
	}
	return out
}

func extractContext(ctx context.Context, r *Record) context.Context {
	return otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(r.Headers))
}
