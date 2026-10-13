package kafka

import (
	"errors"
	"strings"

	"github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/plugin/kotel"
)

// Config описывает подключение к кластеру Kafka.
type Config struct {
	Brokers  []string
	ClientID string
}

func (c Config) validate() error {
	if len(c.Brokers) == 0 {
		return errors.New("kafka: no brokers configured")
	}
	return nil
}

// ParseBrokers разбирает строку вида "kafka:9092,kafka2:9092" (удобно для env).
func ParseBrokers(s string) []string {
	var out []string
	for _, b := range strings.Split(s, ",") {
		if b = strings.TrimSpace(b); b != "" {
			out = append(out, b)
		}
	}
	return out
}

// newTracing подключает OTel к клиенту: franz-go сам кладёт trace context в заголовки
// сообщения при отправке и достаёт его при чтении.
func newTracing() (*kotel.Tracer, kgo.Opt) {
	tracer := kotel.NewTracer()
	k := kotel.NewKotel(kotel.WithTracer(tracer))
	return tracer, kgo.WithHooks(k.Hooks()...)
}
