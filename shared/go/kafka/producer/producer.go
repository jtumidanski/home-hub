// Package producer is a thin retry wrapper around segmentio/kafka-go's Writer.
// It is deliberately small: one Produce call, configurable max attempts,
// structured logging on each retry. No outbox, no partition sentinel — a
// failed final attempt is a logged warning so the caller can decide.
package producer

import (
	"context"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type Writer interface {
	WriteMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type Producer struct {
	writer      Writer
	logger      logrus.FieldLogger
	maxAttempts int
	backoff     time.Duration
}

type Config struct {
	Brokers     []string
	MaxAttempts int
	Backoff     time.Duration
}

func New(cfg Config, logger logrus.FieldLogger) *Producer {
	if cfg.MaxAttempts <= 0 {
		cfg.MaxAttempts = 3
	}
	if cfg.Backoff <= 0 {
		cfg.Backoff = 250 * time.Millisecond
	}
	w := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Brokers...),
		Balancer:     &kafka.Hash{},
		RequiredAcks: kafka.RequireAll,
		Async:        false,
	}
	return &Producer{writer: w, logger: logger, maxAttempts: cfg.MaxAttempts, backoff: cfg.Backoff}
}

func (p *Producer) Produce(ctx context.Context, topic string, key, value []byte, headers map[string]string) error {
	msg := kafka.Message{Topic: topic, Key: key, Value: value}
	for k, v := range headers {
		msg.Headers = append(msg.Headers, kafka.Header{Key: k, Value: []byte(v)})
	}
	var err error
	attempts := p.maxAttempts
	if attempts <= 0 {
		attempts = 1
	}
	for i := 0; i < attempts; i++ {
		err = p.writer.WriteMessages(ctx, msg)
		if err == nil {
			return nil
		}
		p.logger.WithError(err).WithField("topic", topic).WithField("attempt", i+1).Warn("kafka produce failed")
		if i < attempts-1 && p.backoff > 0 {
			time.Sleep(p.backoff)
		}
	}
	return err
}

func (p *Producer) Close() error { return p.writer.Close() }
