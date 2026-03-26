package carrier

import (
	"sync"
	"time"
)

// RateBudget tracks daily API request counts per carrier.
type RateBudget struct {
	mu       sync.Mutex
	counts   map[string]int
	limits   map[string]int
	resetDay int
}

// NewRateBudget creates a budget tracker with per-carrier daily limits.
func NewRateBudget(limits map[string]int) *RateBudget {
	return &RateBudget{
		counts:   make(map[string]int),
		limits:   limits,
		resetDay: time.Now().UTC().YearDay(),
	}
}

// CanRequest returns true if the carrier has remaining budget for today.
func (b *RateBudget) CanRequest(carrierName string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.maybeReset()

	limit, ok := b.limits[carrierName]
	if !ok {
		return true
	}
	return b.counts[carrierName] < limit
}

// Record increments the request count for a carrier.
func (b *RateBudget) Record(carrierName string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.maybeReset()
	b.counts[carrierName]++
}

func (b *RateBudget) maybeReset() {
	today := time.Now().UTC().YearDay()
	if today != b.resetDay {
		b.counts = make(map[string]int)
		b.resetDay = today
	}
}
