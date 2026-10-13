package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"google.golang.org/protobuf/proto"
)

// Message — одно сообщение для пакетной отправки.
type Message struct {
	Topic string
	Key   string
	Value []byte
}

type Producer struct {
	cl *kgo.Client
}

func NewProducer(ctx context.Context, cfg Config) (*Producer, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	_, tracing := newTracing()

	// По умолчанию franz-go использует acks=all и идемпотентный продюсер:
	// повторная отправка при сетевой ошибке не создаёт дубликат в партиции.
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ClientID(cfg.ClientID),
		tracing,
	)
	if err != nil {
		return nil, fmt.Errorf("kafka producer: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := cl.Ping(pingCtx); err != nil {
		cl.Close()
		return nil, fmt.Errorf("kafka producer ping: %w", err)
	}
	return &Producer{cl: cl}, nil
}

// Publish синхронно отправляет сообщение и ждёт подтверждения брокера.
// ctx кладётся в запись, чтобы трейс продолжился у consumer-а.
func (p *Producer) Publish(ctx context.Context, topic, key string, value []byte) error {
	return p.PublishBatch(ctx, []Message{{Topic: topic, Key: key, Value: value}})
}

// PublishBatch отправляет пачку сообщений одним вызовом и ждёт подтверждения всех.
func (p *Producer) PublishBatch(ctx context.Context, msgs []Message) error {
	if len(msgs) == 0 {
		return nil
	}
	recs := make([]*kgo.Record, len(msgs))
	for i, m := range msgs {
		recs[i] = &kgo.Record{Context: ctx, Topic: m.Topic, Key: []byte(m.Key), Value: m.Value}
	}
	return p.cl.ProduceSync(ctx, recs...).FirstErr()
}

// PublishProto сериализует protobuf-сообщение и отправляет его.
func (p *Producer) PublishProto(ctx context.Context, topic, key string, msg proto.Message) error {
	b, err := proto.Marshal(msg)
	if err != nil {
		return fmt.Errorf("marshal %T: %w", msg, err)
	}
	return p.Publish(ctx, topic, key, b)
}

func (p *Producer) Close() {
	p.cl.Close()
}
