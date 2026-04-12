package retention

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRunDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := MigrateRuns(db); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestWriteRun(t *testing.T) {
	db := setupRunDB(t)
	now := time.Now().UTC()
	fin := now.Add(time.Second)
	rec := RunRecord{
		TenantId:   uuid.New(),
		ScopeKind:  ScopeHousehold,
		ScopeId:    uuid.New(),
		Category:   CatProductivityCompletedTasks,
		Trigger:    TriggerScheduled,
		Scanned:    10,
		Deleted:    3,
		StartedAt:  now,
		FinishedAt: &fin,
	}
	if err := WriteRun(context.Background(), db, rec); err != nil {
		t.Fatal(err)
	}

	var got RunEntity
	if err := db.First(&got).Error; err != nil {
		t.Fatal(err)
	}
	if got.Deleted != 3 || got.Category != string(CatProductivityCompletedTasks) {
		t.Errorf("unexpected entity: %+v", got)
	}
}
