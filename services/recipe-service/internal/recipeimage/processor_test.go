package recipeimage

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"testing"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus/hooks/test"
	"gorm.io/gorm"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/objectstore/objectstoretest"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipeimage/imaging"
)

func testJPEG(t testing.TB, w, h int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetRGBA(x, y, color.RGBA{uint8(x), uint8(y), 128, 255})
		}
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

type fixture struct {
	db                    *gorm.DB
	store                 *objectstoretest.Store
	tenantID, householdID uuid.UUID
	proc                  *Processor
}

func newFixture(t *testing.T) fixture {
	db := newTestDB(t)
	store := objectstoretest.New()
	tid, hid := uuid.New(), uuid.New()
	l, _ := test.NewNullLogger()
	return fixture{
		db: db, store: store, tenantID: tid, householdID: hid,
		proc: NewProcessor(l, tenantCtx(tid, hid), db).WithStore(store),
	}
}

func countRows[T any](t *testing.T, db *gorm.DB) int64 {
	t.Helper()
	var n int64
	if err := db.Model(new(T)).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

func TestUploadCreate(t *testing.T) {
	f := newFixture(t)
	rid := uuid.New()
	res, err := f.proc.Upload(rid, testJPEG(t, 2000, 1000))
	if err != nil {
		t.Fatal(err)
	}
	if res.Replaced || len(res.OutboxIDs) != 0 {
		t.Fatalf("first upload: replaced=%v outbox=%v", res.Replaced, res.OutboxIDs)
	}
	m := res.Model
	if m.Width() != 1600 || m.Height() != 800 || m.ContentType() != "image/jpeg" {
		t.Fatalf("model dims %dx%d ct=%s", m.Width(), m.Height(), m.ContentType())
	}
	if !f.store.Has(m.FullKey()) || !f.store.Has(m.ThumbKey()) {
		t.Fatalf("objects not stored: %v", f.store.Keys())
	}
	if countRows[Entity](t, f.db) != 1 {
		t.Fatal("want one image row")
	}
}

func TestUploadReplaceEnqueuesOldKeysAndKeepsCreatedAt(t *testing.T) {
	f := newFixture(t)
	rid := uuid.New()
	first, err := f.proc.Upload(rid, testJPEG(t, 64, 64))
	if err != nil {
		t.Fatal(err)
	}
	second, err := f.proc.Upload(rid, testJPEG(t, 64, 64))
	if err != nil {
		t.Fatal(err)
	}
	if !second.Replaced || len(second.OutboxIDs) != 2 {
		t.Fatalf("replace: replaced=%v outbox=%d", second.Replaced, len(second.OutboxIDs))
	}
	if second.Model.Id() == first.Model.Id() {
		t.Fatal("replace must mint a new image id")
	}
	tu1, fu1 := URLs(rid, first.Model.Id())
	tu2, fu2 := URLs(rid, second.Model.Id())
	if tu1 == tu2 || fu1 == fu2 {
		t.Fatal("replace must change the advertised URLs (cache busting)")
	}
	if !second.Model.CreatedAt().Equal(first.Model.CreatedAt()) {
		t.Fatalf("createdAt not carried forward: %v vs %v", second.Model.CreatedAt(), first.Model.CreatedAt())
	}
	if countRows[Entity](t, f.db) != 1 || countRows[ObjectDeletionEntity](t, f.db) != 2 {
		t.Fatal("want 1 image row and 2 outbox rows")
	}
}

func TestUploadRejectsBadInputWithoutSideEffects(t *testing.T) {
	f := newFixture(t)
	_, err := f.proc.Upload(uuid.New(), []byte("nope"))
	if !errors.Is(err, imaging.ErrUnsupportedFormat) {
		t.Fatalf("err = %v", err)
	}
	if f.store.PutCalls != 0 || countRows[Entity](t, f.db) != 0 {
		t.Fatal("rejected upload must not touch storage or DB")
	}
}

func TestUploadPutFailureCleansUp(t *testing.T) {
	f := newFixture(t)
	f.store.PutHook = func(key string) error {
		if bytes.HasSuffix([]byte(key), []byte("thumb.jpg")) {
			return errors.New("minio down")
		}
		return nil
	}
	_, err := f.proc.Upload(uuid.New(), testJPEG(t, 64, 64))
	if !errors.Is(err, ErrStorage) {
		t.Fatalf("err = %v, want ErrStorage", err)
	}
	if len(f.store.Keys()) != 0 {
		t.Fatalf("full rendition not cleaned up: %v", f.store.Keys())
	}
	if countRows[Entity](t, f.db) != 0 {
		t.Fatal("no row may be written on storage failure")
	}
}

func TestUploadDBFailureCleansUpNewObjects(t *testing.T) {
	f := newFixture(t)
	beforeInsertHook = func(*gorm.DB, *Entity) error { return errors.New("db exploded") }
	t.Cleanup(func() { beforeInsertHook = nil })

	_, err := f.proc.Upload(uuid.New(), testJPEG(t, 64, 64))
	if err == nil || errors.Is(err, ErrStorage) {
		t.Fatalf("err = %v, want a DB error", err)
	}
	if len(f.store.Keys()) != 0 {
		t.Fatalf("new objects not cleaned up: %v", f.store.Keys())
	}
}

// Exercises the retry branch only: the competing row is created inside the
// first transaction, so it rolls back with that transaction.
func TestUploadRetriesOnceOnUniqueViolation(t *testing.T) {
	f := newFixture(t)
	rid := uuid.New()
	calls := 0
	beforeInsertHook = func(tx *gorm.DB, e *Entity) error {
		calls++
		if calls == 1 {
			// Simulate a concurrent first upload committing between our
			// DELETE and INSERT.
			competing := *e
			competing.Id = uuid.New()
			return tx.Create(&competing).Error
		}
		return nil
	}
	t.Cleanup(func() { beforeInsertHook = nil })

	res, err := f.proc.Upload(rid, testJPEG(t, 64, 64))
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Fatalf("hook calls = %d, want 2 (one retry)", calls)
	}
	var rows []Entity
	f.db.Find(&rows)
	if len(rows) != 1 || rows[0].Id != res.Model.Id() {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestDeleteIsIdempotentAndEnqueues(t *testing.T) {
	f := newFixture(t)
	rid := uuid.New()
	if _, err := f.proc.Upload(rid, testJPEG(t, 64, 64)); err != nil {
		t.Fatal(err)
	}
	ids, err := f.proc.Delete(rid)
	if err != nil || len(ids) != 2 {
		t.Fatalf("delete ids=%v err=%v", ids, err)
	}
	ids, err = f.proc.Delete(rid)
	if err != nil || len(ids) != 0 {
		t.Fatalf("second delete ids=%v err=%v", ids, err)
	}
	if _, err := f.proc.GetByRecipeID(rid); !errors.Is(err, ErrImageNotFound) {
		t.Fatalf("get after delete err = %v", err)
	}
}

func TestGetIsHouseholdScoped(t *testing.T) {
	f := newFixture(t)
	rid := uuid.New()
	if _, err := f.proc.Upload(rid, testJPEG(t, 64, 64)); err != nil {
		t.Fatal(err)
	}
	l, _ := test.NewNullLogger()
	other := NewProcessor(l, tenantCtx(f.tenantID, uuid.New()), f.db)
	if _, err := other.GetByRecipeID(rid); !errors.Is(err, ErrImageNotFound) {
		t.Fatalf("other household err = %v", err)
	}
	m, err := other.GetByRecipeIDs([]uuid.UUID{rid})
	if err != nil || len(m) != 0 {
		t.Fatalf("other household batch = %v err=%v", m, err)
	}
	own, err := f.proc.GetByRecipeIDs([]uuid.UUID{rid, uuid.New()})
	if err != nil || len(own) != 1 {
		t.Fatalf("own batch = %v err=%v", own, err)
	}
}

func TestOpenMapsErrors(t *testing.T) {
	tests := []struct {
		name    string
		degrade func(f fixture, m Model)
		want    error
	}{
		{"storage failure", func(f fixture, _ Model) { f.store.GetErr = errors.New("down") }, ErrStorage},
		{"missing object", func(f fixture, m Model) { _ = f.store.Delete(context.Background(), m.FullKey()) }, ErrObjectMissing},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f := newFixture(t)
			res, err := f.proc.Upload(uuid.New(), testJPEG(t, 64, 64))
			if err != nil {
				t.Fatal(err)
			}
			obj, err := f.proc.Open(res.Model, SizeThumb)
			if err != nil {
				t.Fatal(err)
			}
			_ = obj.Body.Close()

			tc.degrade(f, res.Model)
			if _, err := f.proc.Open(res.Model, SizeFull); !errors.Is(err, tc.want) {
				t.Fatalf("err = %v, want %v", err, tc.want)
			}
		})
	}
}

func TestUploadWithoutTenantFails(t *testing.T) {
	db := newTestDB(t)
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, t.Context(), db).WithStore(objectstoretest.New())
	if _, err := p.Upload(uuid.New(), testJPEG(t, 64, 64)); !errors.Is(err, ErrNoTenant) {
		t.Fatalf("err = %v", err)
	}
}
