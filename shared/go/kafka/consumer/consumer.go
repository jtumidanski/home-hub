// Package consumer is a small wrapper around kafka-go's Reader that routes
// each message to a single handler and commits only on success. Consumers
// that want to multiplex by event type can do so inside their handler by
// decoding the envelope's Type field.
package consumer

import (
	"context"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type Reader interface {
	FetchMessage(ctx context.Context) (kafka.Message, error)
	CommitMessages(ctx context.Context, msgs ...kafka.Message) error
	Close() error
}

type Handler func(ctx context.Context, msg kafka.Message) error

type Config struct {
	Brokers []string
	Topic   string
	GroupID string
}

type Manager struct {
	reader  Reader
	logger  logrus.FieldLogger
	handler Handler
}

func New(cfg Config, handler Handler, logger logrus.FieldLogger) *Manager {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: cfg.Brokers,
		Topic:   cfg.Topic,
		GroupID: cfg.GroupID,
	})
	return &Manager{reader: r, logger: logger, handler: handler}
}

func (m *Manager) Run(ctx context.Context) {
	for {
		if ctx.Err() != nil {
			return
		}
		msg, err := m.reader.FetchMessage(ctx)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			m.logger.WithError(err).Warn("kafka fetch failed")
			continue
		}
		if err := m.handler(ctx, msg); err != nil {
			m.logger.WithError(err).WithField("topic", msg.Topic).Error("kafka handler failed; skipping commit")
			continue
		}
		if err := m.reader.CommitMessages(ctx, msg); err != nil {
			m.logger.WithError(err).Warn("kafka commit failed")
		}
	}
}

func (m *Manager) Close() error { return m.reader.Close() }
