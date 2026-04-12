package retention

import (
	"context"
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

// Loop runs fn on a jittered interval (±10%) until ctx is canceled.
// fn receives a derived context and should return promptly when ctx is canceled.
// Loop blocks until ctx is canceled, then returns.
func Loop(ctx context.Context, l logrus.FieldLogger, base time.Duration, fn func(context.Context)) {
	if base <= 0 {
		base = 6 * time.Hour
	}
	l.WithField("base_interval", base.String()).Info("retention loop started")
	for {
		next := jitter(base)
		select {
		case <-ctx.Done():
			l.Info("retention loop stopped")
			return
		case <-time.After(next):
			fn(ctx)
		}
	}
}

// jitter returns d ±10%, never less than half d.
func jitter(d time.Duration) time.Duration {
	delta := float64(d) * 0.10
	offset := (rand.Float64()*2 - 1) * delta
	out := time.Duration(float64(d) + offset)
	if out < d/2 {
		return d / 2
	}
	return out
}
