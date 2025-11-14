package scheduler

import (
	"sync"

	"github.com/google/uuid"
)

// HouseholdTracker tracks which households should be refreshed
type HouseholdTracker struct {
	households map[uuid.UUID]struct{}
	mu         sync.RWMutex
}

// NewHouseholdTracker creates a new household tracker
func NewHouseholdTracker() *HouseholdTracker {
	return &HouseholdTracker{
		households: make(map[uuid.UUID]struct{}),
	}
}

// Track adds a household to be tracked for refresh
func (t *HouseholdTracker) Track(householdID uuid.UUID) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.households[householdID] = struct{}{}
}

// Untrack removes a household from tracking
func (t *HouseholdTracker) Untrack(householdID uuid.UUID) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.households, householdID)
}

// IsTracked returns true if the household is being tracked
func (t *HouseholdTracker) IsTracked(householdID uuid.UUID) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	_, exists := t.households[householdID]
	return exists
}

// GetAll returns all tracked households
func (t *HouseholdTracker) GetAll() []uuid.UUID {
	t.mu.RLock()
	defer t.mu.RUnlock()

	households := make([]uuid.UUID, 0, len(t.households))
	for id := range t.households {
		households = append(households, id)
	}

	return households
}

// Count returns the number of tracked households
func (t *HouseholdTracker) Count() int {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return len(t.households)
}

// Clear removes all tracked households
func (t *HouseholdTracker) Clear() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.households = make(map[uuid.UUID]struct{})
}
