package recipeimage

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus/hooks/test"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/objectstore"
)

func TestReaperDrainsOutbox(t *testing.T) {
	f := newFixture(t)
	rid := uuid.New()
	first, _ := f.proc.Upload(rid, testJPEG(t, 64, 64))
	second, err := f.proc.Upload(rid, testJPEG(t, 64, 64))
	if err != nil {
		t.Fatal(err)
	}
	l, _ := test.NewNullLogger()
	r := NewObjectReaper(l, f.db, f.store)

	r.DrainKeys(context.Background(), second.OutboxIDs)

	if f.store.Has(first.Model.FullKey()) || f.store.Has(first.Model.ThumbKey()) {
		t.Fatal("old objects not deleted")
	}
	if !f.store.Has(second.Model.FullKey()) {
		t.Fatal("current object deleted")
	}
	if countRows[ObjectDeletionEntity](t, f.db) != 0 {
		t.Fatal("outbox rows not removed")
	}
}

func TestReaperFailureIncrementsAttempts(t *testing.T) {
	f := newFixture(t)
	rid := uuid.New()
	_, _ = f.proc.Upload(rid, testJPEG(t, 64, 64))
	if _, err := f.proc.Delete(rid); err != nil {
		t.Fatal(err)
	}
	f.store.DeleteErr = errors.New("down")
	l, _ := test.NewNullLogger()
	r := NewObjectReaper(l, f.db, f.store)

	r.Run(context.Background())

	var rows []ObjectDeletionEntity
	f.db.Find(&rows)
	if len(rows) != 2 {
		t.Fatalf("rows = %d, want 2", len(rows))
	}
	for _, row := range rows {
		if row.Attempts != 1 || row.LastError == nil || *row.LastError == "" {
			t.Fatalf("row not marked failed: %+v", row)
		}
	}

	f.store.DeleteErr = nil
	r.Run(context.Background())
	if countRows[ObjectDeletionEntity](t, f.db) != 0 {
		t.Fatal("retry did not drain")
	}
}

func TestReaperTreatsNotFoundAsSuccess(t *testing.T) {
	f := newFixture(t)
	rid := uuid.New()
	_, _ = f.proc.Upload(rid, testJPEG(t, 64, 64))
	_, _ = f.proc.Delete(rid)
	f.store.DeleteErr = objectstore.ErrNotFound
	l, _ := test.NewNullLogger()
	NewObjectReaper(l, f.db, f.store).Run(context.Background())
	if countRows[ObjectDeletionEntity](t, f.db) != 0 {
		t.Fatal("ErrNotFound must count as success")
	}
}
