package kafka

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"market/shared/infra/logger"
)

type Handler func(ctx context.Context, rec *Record) error

type BatchHandler func(ctx context.Context, recs []*Record) error

func Permanent(err error) error { return permanentError{err: err} }

type permanentError struct{ err error }

func (e permanentError) Error() string { return e.err.Error() }
func (e permanentError) Unwrap() error { return e.err }

func isPermanent(err error) bool {
	var p permanentError
	return errors.As(err, &p)
}

const maxBatch = 100

const (
	hdrPrefix        = "x-dlq-"
	hdrOrigTopic     = "x-dlq-original-topic"
	hdrOrigPartition = "x-dlq-original-partition"
	hdrOrigOffset    = "x-dlq-original-offset"
	hdrFailures      = "x-dlq-failures"
	hdrError         = "x-dlq-error"
	hdrGroup         = "x-dlq-group"
	hdrFailedAt      = "x-dlq-failed-at"
)

type consumerOptions struct {
	dlqTopic    string
	maxAttempts int
	delay       time.Duration
}

type ConsumerOption func(*consumerOptions)

func WithDLQ(topic string) ConsumerOption { return func(o *consumerOptions) { o.dlqTopic = topic } }

func WithMaxAttempts(n int) ConsumerOption { return func(o *consumerOptions) { o.maxAttempts = n } }

func WithDelay(d time.Duration) ConsumerOption { return func(o *consumerOptions) { o.delay = d } }

type Consumer struct {
	cg      sarama.ConsumerGroup
	groupID string
	topics  []string
	tracer  trace.Tracer
	log     *zap.Logger
	opts    consumerOptions
	dlq     *Producer
}

func NewConsumer(cfg Config, group string, topics []string, log *zap.Logger, opts ...ConsumerOption) (*Consumer, error) {
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	o := consumerOptions{maxAttempts: 5}
	for _, opt := range opts {
		opt(&o)
	}

	sc := newSaramaConfig(cfg)
	sc.Consumer.Offsets.Initial = sarama.OffsetOldest
	sc.Consumer.Offsets.AutoCommit.Enable = false
	sc.Consumer.Return.Errors = true

	cg, err := sarama.NewConsumerGroup(cfg.Brokers, group, sc)
	if err != nil {
		return nil, fmt.Errorf("kafka consumer: %w", err)
	}
	c := &Consumer{
		cg: cg, groupID: group, topics: topics,
		tracer: otel.Tracer("market/shared/kafka"), log: log, opts: o,
	}
	if o.dlqTopic != "" {
		p, err := NewProducer(context.Background(), cfg)
		if err != nil {
			cg.Close()
			return nil, fmt.Errorf("kafka consumer: dlq producer: %w", err)
		}
		c.dlq = p
	}

	go func() {
		for err := range cg.Errors() {
			log.Error("kafka consumer group error", zap.Error(err))
		}
	}()
	return c, nil
}

func (c *Consumer) Run(ctx context.Context, h Handler) error {
	return c.RunBatch(ctx, func(ctx context.Context, recs []*Record) error {
		for _, r := range recs {
			if err := c.callOne(ctx, r, h); err != nil {
				return err
			}
		}
		return nil
	})
}

func (c *Consumer) callOne(ctx context.Context, r *Record, h Handler) error {
	rctx, span := c.tracer.Start(extractContext(ctx, r), r.Topic+" process",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithAttributes(
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.destination.name", r.Topic),
			attribute.Int64("messaging.kafka.partition", int64(r.Partition)),
			attribute.Int64("messaging.kafka.offset", r.Offset),
		))
	defer span.End()
	err := h(rctx, r)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return err
}

func (c *Consumer) RunBatch(ctx context.Context, h BatchHandler) error {
	gh := &groupHandler{c: c, h: h}
	for {
		if err := c.cg.Consume(ctx, c.topics, gh); err != nil {
			if errors.Is(err, sarama.ErrClosedConsumerGroup) || ctx.Err() != nil {
				return nil
			}
			c.log.Error("kafka consume failed", zap.Error(err))
			if !sleep(ctx, time.Second) {
				return nil
			}
		}
		if ctx.Err() != nil {
			return nil
		}
	}
}

type groupHandler struct {
	c *Consumer
	h BatchHandler
}

func (*groupHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (*groupHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (g *groupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	ctx := sess.Context()
	for {
		msgs, ok := nextBatch(ctx, claim.Messages())
		if !ok {
			return nil
		}
		if d := g.c.opts.delay; d > 0 {
			if !sleepUntil(ctx, msgs[len(msgs)-1].Timestamp.Add(d)) {
				return nil
			}
		}
		recs := make([]*Record, len(msgs))
		for i, m := range msgs {
			recs[i] = toRecord(m)
		}
		if !g.c.handleBatch(ctx, recs, g.h) {
			return nil
		}
		sess.MarkMessage(msgs[len(msgs)-1], "")
		sess.Commit()
	}
}

func nextBatch(ctx context.Context, ch <-chan *sarama.ConsumerMessage) ([]*sarama.ConsumerMessage, bool) {
	select {
	case <-ctx.Done():
		return nil, false
	case m, ok := <-ch:
		if !ok {
			return nil, false
		}
		batch := []*sarama.ConsumerMessage{m}
		for len(batch) < maxBatch {
			select {
			case m, ok := <-ch:
				if !ok {
					return batch, true
				}
				batch = append(batch, m)
			default:
				return batch, true
			}
		}
		return batch, true
	}
}

func (c *Consumer) handleBatch(ctx context.Context, recs []*Record, h BatchHandler) bool {
	links := make([]trace.Link, 0, len(recs))
	for _, r := range recs {
		sc := trace.SpanContextFromContext(extractContext(context.Background(), r))
		if sc.IsValid() {
			links = append(links, trace.Link{SpanContext: sc})
		}
	}
	ctx, span := c.tracer.Start(ctx, "kafka.ProcessBatch",
		trace.WithSpanKind(trace.SpanKindConsumer),
		trace.WithLinks(links...),
		trace.WithAttributes(
			attribute.Int("kafka.batch_size", len(recs)),
			attribute.String("kafka.topic", recs[0].Topic),
			attribute.Int64("kafka.partition", int64(recs[0].Partition)),
		))
	defer span.End()
	return c.process(ctx, recs, h)
}

type outcome int

const (
	outcomeOK outcome = iota
	outcomeFailed
	outcomeStopped
)

func (c *Consumer) process(ctx context.Context, recs []*Record, h BatchHandler) bool {
	res, err := c.runWithRetry(ctx, recs, h)
	switch res {
	case outcomeOK:
		return true
	case outcomeStopped:
		return false
	}
	if isPermanent(err) && len(recs) > 1 {
		mid := len(recs) / 2
		return c.process(ctx, recs[:mid], h) && c.process(ctx, recs[mid:], h)
	}
	return c.deadLetter(ctx, recs, err)
}

func (c *Consumer) runWithRetry(ctx context.Context, recs []*Record, h BatchHandler) (outcome, error) {
	limit := 0
	if c.dlq != nil {
		limit = c.opts.maxAttempts
	}
	log := logger.WithTrace(ctx, c.log).With(
		zap.String("topic", recs[0].Topic), zap.Int32("partition", recs[0].Partition),
		zap.Int("records", len(recs)))

	for attempt := 0; ; attempt++ {
		err := h(ctx, recs)
		if err == nil {
			return outcomeOK, nil
		}
		if isPermanent(err) {
			return outcomeFailed, err
		}
		if limit > 0 && attempt+1 >= limit {
			log.Warn("kafka: retries exhausted", zap.Int("attempts", attempt+1), zap.Error(err))
			return outcomeFailed, err
		}
		log.Warn("kafka: handler failed, will retry", zap.Int("attempt", attempt), zap.Error(err))
		if !sleep(ctx, backoff(attempt)) {
			return outcomeStopped, nil
		}
	}
}

func (c *Consumer) deadLetter(ctx context.Context, recs []*Record, cause error) bool {
	log := logger.WithTrace(ctx, c.log)
	if c.dlq == nil {
		for _, r := range recs {
			log.Error("kafka: skipping unprocessable record (no DLQ configured)",
				zap.String("topic", r.Topic), zap.Int32("partition", r.Partition),
				zap.Int64("offset", r.Offset), zap.Error(cause))
		}
		return true
	}

	msgs := make([]Message, len(recs))
	for i, r := range recs {
		msgs[i] = c.dlqMessage(r, cause)
	}
	for attempt := 0; ; attempt++ {
		err := c.dlq.PublishBatch(ctx, msgs)
		if err == nil {
			for _, r := range recs {
				log.Warn("kafka: record moved to DLQ",
					zap.String("dlq", c.opts.dlqTopic), zap.String("topic", r.Topic),
					zap.Int32("partition", r.Partition), zap.Int64("offset", r.Offset),
					zap.NamedError("cause", cause))
			}
			return true
		}
		log.Error("kafka: DLQ publish failed, will retry", zap.Int("attempt", attempt), zap.Error(err))
		if !sleep(ctx, backoff(attempt)) {
			return false
		}
	}
}

func (c *Consumer) dlqMessage(r *Record, cause error) Message {
	h := make(map[string]string, 8)
	for k, v := range r.Headers {
		if strings.HasPrefix(k, hdrPrefix) {
			h[k] = v
		}
	}
	if _, ok := h[hdrOrigTopic]; !ok {
		h[hdrOrigTopic] = r.Topic
		h[hdrOrigPartition] = strconv.Itoa(int(r.Partition))
		h[hdrOrigOffset] = strconv.FormatInt(r.Offset, 10)
	}
	failures := 1
	if n, err := strconv.Atoi(h[hdrFailures]); err == nil {
		failures = n + 1
	}
	msg := cause.Error()
	if len(msg) > 512 {
		msg = msg[:512]
	}
	h[hdrFailures] = strconv.Itoa(failures)
	h[hdrError] = msg
	h[hdrGroup] = c.groupID
	h[hdrFailedAt] = time.Now().UTC().Format(time.RFC3339Nano)

	return Message{Topic: c.opts.dlqTopic, Key: string(r.Key), Value: r.Value, Headers: h}
}

func backoff(attempt int) time.Duration {
	return 100 * time.Millisecond << min(attempt, 6)
}

func sleep(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

func sleepUntil(ctx context.Context, t time.Time) bool {
	d := time.Until(t)
	if d <= 0 {
		return true
	}
	return sleep(ctx, d)
}

func (c *Consumer) Close() error {
	err := c.cg.Close()
	if c.dlq != nil {
		err = errors.Join(err, c.dlq.Close())
	}
	return err
}

func Proto[M proto.Message](newMsg func() M, fn func(ctx context.Context, msg M) error) Handler {
	return func(ctx context.Context, rec *Record) error {
		msg := newMsg()
		if err := proto.Unmarshal(rec.Value, msg); err != nil {
			return Permanent(fmt.Errorf("unmarshal %T: %w", msg, err))
		}
		return fn(ctx, msg)
	}
}
