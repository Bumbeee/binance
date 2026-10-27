package kafka

import (
	"errors"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

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

func ParseBrokers(s string) []string {
	var out []string
	for _, b := range strings.Split(s, ",") {
		if b = strings.TrimSpace(b); b != "" {
			out = append(out, b)
		}
	}
	return out
}

func newSaramaConfig(cfg Config) *sarama.Config {
	sc := sarama.NewConfig()
	sc.Version = sarama.V3_6_0_0
	sc.ClientID = cfg.ClientID
	sc.Net.DialTimeout = 5 * time.Second
	return sc
}
