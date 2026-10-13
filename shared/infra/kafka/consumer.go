package kafka

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kotel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"market/shared/infra/logger"
	"market/shared/tracing"
)

// Handler обрабатывает одно сообщение.
type Handler func(ctx context.Context, rec *kgo.Record) error

// BatchHandler обрабатывает пачку сообщений, полученную за один poll.
type BatchHandler func(ctx context.Context, recs []*kgo.Record) error

// Permanent помечает ошибку как неисправимую (например, битые данные): повтор не поможет,
// сообщение будет пропущено. Любая другая ошибка считается временной и повторяется с паузой.
func Permanent(err error) error { return permanentError{err: err} }

type permanentError struct{ err error }

func (e permanentError) Error() string { return e.err.Error() }
func (e permanentError) Unwrap() error { return e.err }

type Consumer struct {
	cl     *kgo.Client
	tracer *kotel.Tracer
	log    *zap.Logger
}

func NewConsumer(cfg Config, group string, topics []string, log *zap.Logger) (*Consumer, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	tracer, tracing := newTracing()

	cl, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ClientID(cfg.ClientID),
		kgo.ConsumerGroup(group),
		kgo.ConsumeTopics(topics...),
		// Новая группа читает топик с начала.
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()),
		// Офсет коммитим сами, только после успешной обработки.
		kgo.DisableAutoCommit(),
		tracing,
	)
	if err != nil {
		return nil, fmt.Errorf("kafka consumer: %w", err)
	}
	return &Consumer{cl: cl, tracer: tracer, log: log}, nil
}

// Run читает сообщения по одному до отмены ctx. Доставка at-least-once: после ребаланса
// или падения часть сообщений может прийти повторно, поэтому handler должен быть идемпотентным.
func (c *Consumer) Run(ctx context.Context, h Handler) error {
	return c.loop(ctx, func(recs []*kgo.Record) bool {
		for _, rec := range recs {
			if !c.processOne(ctx, rec, h) {
				return false
			}
		}
		return true
	})
}

// RunBatch читает сообщения пачками: handler получает всё, что вернул один poll.
func (c *Consumer) RunBatch(ctx context.Context, h BatchHandler) error {
	return c.loop(ctx, func(recs []*kgo.Record) bool {
		return c.processBatch(ctx, recs, h)
	})
}

// loop — общий цикл: poll → обработка → коммит. process возвращает false, если ctx отменён.
func (c *Consumer) loop(ctx context.Context, process func(recs []*kgo.Record) bool) error {
	for {
		fetches := c.cl.PollFetches(ctx)
		if fetches.IsClientClosed() || ctx.Err() != nil {
			return nil
		}
		fetches.EachError(func(topic string, partition int32, err error) {
			c.log.Error("kafka fetch error",
				zap.String("topic", topic), zap.Int32("partition", partition), zap.Error(err))
		})

		var recs []*kgo.Record
		fetches.EachRecord(func(r *kgo.Record) {
			recs = append(recs, r)
		})

		if len(recs) > 0 && !process(recs) {
			return nil
		}
		if err := c.cl.CommitUncommittedOffsets(ctx); err != nil && ctx.Err() == nil {
			c.log.Error("kafka commit failed", zap.Error(err))
		}
	}
}

func (c *Consumer) processOne(ctx context.Context, rec *kgo.Record, h Handler) bool {
	// Span "process" продолжает трейс продюсера (контекст из заголовков записи).
	rctx, span := c.tracer.WithProcessSpan(rec)
	defer span.End()

	log := logger.WithTrace(rctx, c.log).With(
		zap.String("topic", rec.Topic),
		zap.Int32("partition", rec.Partition),
		zap.Int64("offset", rec.Offset),
	)

	for attempt := 0; ; attempt++ {
		err := h(rctx, rec)
		if err == nil {
			return true
		}

		span.RecordError(err)
		var perm permanentError
		if errors.As(err, &perm) {
			span.SetStatus(codes.Error, err.Error())
			log.Error("kafka: skipping unprocessable record", zap.Error(err))
			return true
		}

		log.Warn("kafka: handler failed, will retry", zap.Int("attempt", attempt), zap.Error(err))
		if !sleep(ctx, backoff(attempt)) {
			return false
		}
	}
}

func (c *Consumer) processBatch(ctx context.Context, recs []*kgo.Record, h BatchHandler) bool {
	ctx, span := tracing.Start(ctx, "kafka.ProcessBatch",
		attribute.Int("kafka.batch_size", len(recs)),
		attribute.String("kafka.topic", recs[0].Topic),
	)
	var err error
	defer span.End(&err)

	for attempt := 0; ; attempt++ {
		if err = h(ctx, recs); err == nil {
			return true
		}
		var perm permanentError
		if errors.As(err, &perm) {
			logger.WithTrace(ctx, c.log).Error("kafka: skipping unprocessable batch", zap.Error(err))
			return true
		}
		logger.WithTrace(ctx, c.log).Warn("kafka: batch failed, will retry",
			zap.Int("attempt", attempt), zap.Error(err))
		if !sleep(ctx, backoff(attempt)) {
			return false
		}
	}
}

func backoff(attempt int) time.Duration {
	return 100 * time.Millisecond << min(attempt, 6) // до ~6.4 с
}

// sleep ждёт d или отмены ctx; false значит ctx отменён.
func sleep(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

func (c *Consumer) Close() {
	c.cl.Close()
}

// Proto превращает типизированный обработчик в Handler: разбирает value в protobuf.
// Битое сообщение помечается как Permanent.
func Proto[M proto.Message](newMsg func() M, fn func(ctx context.Context, msg M) error) Handler {
	return func(ctx context.Context, rec *kgo.Record) error {
		msg := newMsg()
		if err := proto.Unmarshal(rec.Value, msg); err != nil {
			return Permanent(fmt.Errorf("unmarshal %T: %w", msg, err))
		}
		return fn(ctx, msg)
	}
}
