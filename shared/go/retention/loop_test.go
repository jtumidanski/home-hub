package retention

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus/hooks/test"
)

func TestLoopRunsAndStops(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var calls int32
	l, _ := test.NewNullLogger()
	done := make(chan struct{})
	go func() {
		Loop(ctx, l, 10*time.Millisecond, func(ctx context.Context) {
			atomic.AddInt32(&calls, 1)
		})
		close(done)
	}()

	time.Sleep(60 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Loop did not exit on cancel")
	}
	if atomic.LoadInt32(&calls) == 0 {
		t.Error("expected at least one call")
	}
}

func TestJitterWithinBounds(t *testing.T) {
	for i := 0; i < 200; i++ {
		j := jitter(time.Hour)
		if j < time.Hour/2 || j > time.Hour+time.Hour/10+time.Second {
			t.Errorf("jitter out of bounds: %v", j)
		}
	}
}
