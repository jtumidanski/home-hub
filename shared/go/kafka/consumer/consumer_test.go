package consumer

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/sirupsen/logrus"
)

type fakeReader struct {
	mu     sync.Mutex
	msgs   []kafka.Message
	idx    int
	done   chan struct{}
	closer sync.Once
}

func (f *fakeReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.idx >= len(f.msgs) {
		f.closer.Do(func() { close(f.done) })
		return kafka.Message{}, errors.New("EOF")
	}
	m := f.msgs[f.idx]
	f.idx++
	return m, nil
}
func (f *fakeReader) CommitMessages(_ context.Context, _ ...kafka.Message) error { return nil }
func (f *fakeReader) Close() error                                                { return nil }

func TestDispatchInvokesHandler(t *testing.T) {
	got := make(chan kafka.Message, 1)
	r := &fakeReader{msgs: []kafka.Message{{Topic: "t", Value: []byte(`x`)}}, done: make(chan struct{})}
	l := logrus.New()
	m := &Manager{reader: r, logger: l, handler: func(ctx context.Context, msg kafka.Message) error { got <- msg; return nil }}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go m.Run(ctx)
	select {
	case msg := <-got:
		if msg.Topic != "t" {
			t.Fatalf("unexpected topic %q", msg.Topic)
		}
	case <-time.After(time.Second):
		t.Fatal("handler never invoked")
	}
}
