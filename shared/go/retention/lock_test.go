package retention

import (
	"testing"

	"github.com/google/uuid"
)

func TestLockKeyDeterministic(t *testing.T) {
	id := uuid.New()
	a := LockKey(id, CatProductivityCompletedTasks)
	b := LockKey(id, CatProductivityCompletedTasks)
	if a != b {
		t.Errorf("LockKey not deterministic: %d != %d", a, b)
	}
}

func TestLockKeyDifferentiation(t *testing.T) {
	id1 := uuid.New()
	id2 := uuid.New()
	if LockKey(id1, CatProductivityCompletedTasks) == LockKey(id2, CatProductivityCompletedTasks) {
		t.Error("different tenants produced identical keys")
	}
	if LockKey(id1, CatProductivityCompletedTasks) == LockKey(id1, CatRecipeRestorationAudit) {
		t.Error("different categories produced identical keys")
	}
}
