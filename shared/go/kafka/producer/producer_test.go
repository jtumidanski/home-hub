package producer

import (
	"context"
	"errors"
	"testing"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type stubWriter struct {
	msgs []kafka.Message
	err  error
}

func (s *stubWriter) WriteMessages(_ context.Context, m ...kafka.Message) error {
	s.msgs = append(s.msgs, m...)
	return s.err
}

func (s *stubWriter) Close() error { return nil }

func TestProduceWritesMessage(t *testing.T) {
	sw := &stubWriter{}
	l := logrus.New()
	p := &Producer{writer: sw, logger: l}
	err := p.Produce(context.Background(), "topic", []byte("k"), []byte(`{"x":1}`), nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(sw.msgs) != 1 || sw.msgs[0].Topic != "topic" {
		t.Fatalf("unexpected: %+v", sw.msgs)
	}
}

func TestProduceRetriesThenFails(t *testing.T) {
	sw := &stubWriter{err: errors.New("boom")}
	l := logrus.New()
	p := &Producer{writer: sw, logger: l, maxAttempts: 3}
	err := p.Produce(context.Background(), "topic", []byte("k"), []byte("v"), nil)
	if err == nil {
		t.Fatal("expected error")
	}
	if len(sw.msgs) != 3 {
		t.Fatalf("expected 3 attempts, got %d", len(sw.msgs))
	}
}
