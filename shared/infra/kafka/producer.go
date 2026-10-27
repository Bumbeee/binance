package kafka

import (
	"context"
	"errors"
	"fmt"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/protobuf/proto"
)

type Message struct {
	Topic   string
	Key     string
	Value   []byte
	Headers map[string]string
}

type Producer struct {
	sp     sarama.SyncProducer
	tracer trace.Tracer
}

func NewProducer(ctx context.Context, cfg Config) (*Producer, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	sc := newSaramaConfig(cfg)
	sc.Producer.RequiredAcks = sarama.WaitForAll
	sc.Producer.Return.Successes = true
	sc.Producer.Return.Errors = true
	sc.Producer.Idempotent = true
	sc.Net.MaxOpenRequests = 1
	sc.Producer.Partitioner = sarama.NewReferenceHashPartitioner

	sp, err := sarama.NewSyncProducer(cfg.Brokers, sc)
	if err != nil {
		return nil, fmt.Errorf("kafka producer: %w", err)
	}
	return &Producer{sp: sp, tracer: otel.Tracer("market/shared/kafka")}, nil
}

func (p *Producer) Publish(ctx context.Context, m Message) error {
	return p.PublishBatch(ctx, []Message{m})
}

func (p *Producer) PublishProto(ctx context.Context, topic, key string, msg proto.Message) error {
	b, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal %T: %w", msg, err)
	}
	return p.Publish(ctx, Message{Topic: topic, Key: key, Value: b})
}

func (p *Producer) PublishBatch(ctx context.Context, msgs []Message) error {
	if len(msgs) == 0 {
		return nil
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	pms := make([]*sarama.ProducerMessage, len(msgs))
	spans := make([]trace.Span, len(msgs))
	for i, m := range msgs {
		sctx, span := p.tracer.Start(ctx, m.Topic+" publish",
			trace.WithSpanKind(trace.SpanKindProducer),
			trace.WithAttributes(
				attribute.String("messaging.system", "kafka"),
				attribute.String("messaging.destination.name", m.Topic),
				attribute.String("messaging.kafka.message.key", m.Key),
			))
		spans[i] = span

		hdrs := injectHeaders(sctx)
		for k, v := range m.Headers {
			hdrs = append(hdrs, sarama.RecordHeader{Key: []byte(k), Value: []byte(v)})
		}
		pm := &sarama.ProducerMessage{
			Topic:   m.Topic,
			Value:   sarama.ByteEncoder(m.Value),
			Headers: hdrs,
		}

		if m.Key != "" {
			pm.Key = sarama.StringEncoder(m.Key)
		}
		pms[i] = pm
	}

	err := p.sp.SendMessages(pms)

	failed := map[*sarama.ProducerMessage]error{}
	var perrs sarama.ProducerErrors
	if errors.As(err, &perrs) {
		for _, pe := range perrs {
			failed[pe.Msg] = pe.Err
		}
	}
	for i, span := range spans {
		if e, bad := failed[pms[i]]; bad {
			span.RecordError(e)
			span.SetStatus(codes.Error, e.Error())
		} else if err != nil && len(failed) == 0 {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}
	return err
}

func (p *Producer) Close() error { return p.sp.Close() }
