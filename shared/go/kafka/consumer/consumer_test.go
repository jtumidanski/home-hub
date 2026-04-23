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
	mu      sync.Mutex
	msgs    []kafka.Message
	idx     int
	commits int
	done    chan struct{}
	closer  sync.Once
	// block: if true, FetchMessage blocks until ctx is cancelled.
	block bool
}

func (f *fakeReader) FetchMessage(ctx context.Context) (kafka.Message, error) {
	if f.block {
		<-ctx.Done()
		return kafka.Message{}, ctx.Err()
	}
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
func (f *fakeReader) CommitMessages(_ context.Context, _ ...kafka.Message) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.commits++
	return nil
}
func (f *fakeReader) Close() error { return nil }

func (f *fakeReader) commitCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.commits
}

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

func TestCommitOnHandlerSuccess(t *testing.T) {
	handled := make(chan struct{}, 1)
	r := &fakeReader{msgs: []kafka.Message{{Topic: "t", Value: []byte(`x`)}}, done: make(chan struct{})}
	l := logrus.New()
	m := &Manager{reader: r, logger: l, handler: func(ctx context.Context, msg kafka.Message) error {
		handled <- struct{}{}
		return nil
	}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go m.Run(ctx)
	select {
	case <-handled:
	case <-time.After(time.Second):
		t.Fatal("handler never invoked")
	}
	// Wait for Run to finish processing (EOF path) so commit is observable.
	select {
	case <-r.done:
	case <-time.After(time.Second):
		t.Fatal("reader never reached EOF")
	}
	// Give the loop a moment to call CommitMessages before we check.
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if r.commitCount() == 1 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("expected 1 commit, got %d", r.commitCount())
}

func TestSkipCommitOnHandlerError(t *testing.T) {
	r := &fakeReader{msgs: []kafka.Message{{Topic: "t", Value: []byte(`x`)}}, done: make(chan struct{})}
	l := logrus.New()
	m := &Manager{reader: r, logger: l, handler: func(ctx context.Context, msg kafka.Message) error {
		return errors.New("handler boom")
	}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go m.Run(ctx)
	select {
	case <-r.done:
	case <-time.After(time.Second):
		t.Fatal("reader never reached EOF")
	}
	// Let any stray commit happen (shouldn't).
	time.Sleep(50 * time.Millisecond)
	if got := r.commitCount(); got != 0 {
		t.Fatalf("expected 0 commits on handler error, got %d", got)
	}
}

func TestGracefulShutdownOnCtxCancel(t *testing.T) {
	r := &fakeReader{block: true, done: make(chan struct{})}
	l := logrus.New()
	m := &Manager{reader: r, logger: l, handler: func(ctx context.Context, msg kafka.Message) error { return nil }}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		m.Run(ctx)
		close(done)
	}()
	cancel()
	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Run did not return within 100ms of ctx cancel")
	}
}
