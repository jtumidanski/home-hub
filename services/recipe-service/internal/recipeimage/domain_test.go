package recipeimage

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

func seedImage(t *testing.T, db *gorm.DB, tenantID, householdID, recipeID uuid.UUID) Entity {
	t.Helper()
	id := uuid.New()
	full, thumb := objectKeys(tenantID, householdID, recipeID, id)
	now := time.Now().UTC()
	e := Entity{
		Id: id, TenantId: tenantID, HouseholdId: householdID, RecipeId: recipeID,
		FullKey: full, ThumbKey: thumb, ContentType: "image/jpeg",
		Width: 1600, Height: 900, FullBytes: 1000, ThumbBytes: 100,
		CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatal(err)
	}
	return e
}

func TestObjectKeysAndURLs(t *testing.T) {
	tid, hid, rid, iid := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	full, thumb := objectKeys(tid, hid, rid, iid)
	prefix := tid.String() + "/" + hid.String() + "/recipes/" + rid.String() + "/" + iid.String() + "/"
	if full != prefix+"full.jpg" || thumb != prefix+"thumb.jpg" {
		t.Fatalf("keys = %q, %q", full, thumb)
	}
	tu, fu := URLs(rid, iid)
	if tu != "/api/v1/recipes/"+rid.String()+"/image?size=thumb&v="+iid.String() {
		t.Fatalf("thumb url = %q", tu)
	}
	if fu != "/api/v1/recipes/"+rid.String()+"/image?size=full&v="+iid.String() {
		t.Fatalf("full url = %q", fu)
	}
}

func TestParseSize(t *testing.T) {
	for in, want := range map[string]Size{"": SizeFull, "full": SizeFull, "thumb": SizeThumb} {
		got, ok := ParseSize(in)
		if !ok || got != want {
			t.Errorf("ParseSize(%q) = %q,%v", in, got, ok)
		}
	}
	for _, bad := range []string{"FULL", "medium", "thumb "} {
		if _, ok := ParseSize(bad); ok {
			t.Errorf("ParseSize(%q) accepted", bad)
		}
	}
}

func TestBuilderValidation(t *testing.T) {
	if _, err := NewBuilder().SetFullKey("a").SetThumbKey("b").Build(); !errors.Is(err, ErrRecipeIDRequired) {
		t.Fatalf("err = %v", err)
	}
	if _, err := NewBuilder().SetRecipeID(uuid.New()).Build(); !errors.Is(err, ErrKeysRequired) {
		t.Fatalf("err = %v", err)
	}
}

func TestMakeAndKey(t *testing.T) {
	db := newTestDB(t)
	e := seedImage(t, db, uuid.New(), uuid.New(), uuid.New())
	m, err := Make(e)
	if err != nil {
		t.Fatal(err)
	}
	if m.Key(SizeFull) != e.FullKey || m.Key(SizeThumb) != e.ThumbKey || m.Width() != 1600 {
		t.Fatalf("model mismatch: %+v", m)
	}
}

func TestProviderScopesByHousehold(t *testing.T) {
	db := newTestDB(t)
	tid, hidA, hidB := uuid.New(), uuid.New(), uuid.New()
	r1, r2 := uuid.New(), uuid.New()
	seedImage(t, db, tid, hidA, r1)
	seedImage(t, db, tid, hidB, r2)

	ctxA := tenantCtx(tid, hidA)
	if _, err := getByRecipeID(r1, hidA)(db.WithContext(ctxA))(); err != nil {
		t.Fatalf("own household: %v", err)
	}
	if _, err := getByRecipeID(r2, hidA)(db.WithContext(ctxA))(); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("other household err = %v, want not found", err)
	}
	rows, err := getByRecipeIDs([]uuid.UUID{r1, r2}, hidA)(db.WithContext(ctxA))
	if err != nil || len(rows) != 1 || rows[0].RecipeId != r1 {
		t.Fatalf("batched rows = %+v err=%v", rows, err)
	}
	// Tenant callback: another tenant sees nothing even with the right household.
	if _, err := getByRecipeID(r1, hidA)(db.WithContext(tenantCtx(uuid.New(), hidA)))(); !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("other tenant err = %v", err)
	}
}

func TestDeleteReturningEnqueuesKeys(t *testing.T) {
	db := newTestDB(t)
	tid, hid, rid := uuid.New(), uuid.New(), uuid.New()
	e := seedImage(t, db, tid, hid, rid)

	var ids []uuid.UUID
	err := db.WithContext(tenantCtx(tid, hid)).Transaction(func(tx *gorm.DB) error {
		removed, err := deleteByRecipeIDReturning(tx, rid, hid)
		if err != nil {
			return err
		}
		if len(removed) != 1 || removed[0].FullKey != e.FullKey {
			t.Fatalf("removed = %+v", removed)
		}
		ids, err = enqueueObjectDeletions(tx, removed)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != 2 {
		t.Fatalf("outbox ids = %d, want 2", len(ids))
	}
	var keys []string
	db.Model(&ObjectDeletionEntity{}).Order("object_key").Pluck("object_key", &keys)
	if len(keys) != 2 || !strings.HasSuffix(keys[0], "full.jpg") || !strings.HasSuffix(keys[1], "thumb.jpg") {
		t.Fatalf("outbox keys = %v", keys)
	}
	var count int64
	db.Model(&Entity{}).Count(&count)
	if count != 0 {
		t.Fatalf("image rows = %d, want 0", count)
	}
}

func TestPurgeByRecipeIDsTx(t *testing.T) {
	db := newTestDB(t)
	tid, hid := uuid.New(), uuid.New()
	r1, r2, r3 := uuid.New(), uuid.New(), uuid.New()
	seedImage(t, db, tid, hid, r1)
	seedImage(t, db, tid, hid, r2)
	seedImage(t, db, tid, hid, r3)

	n, err := PurgeByRecipeIDsTx(db, []uuid.UUID{r1, r2})
	if err != nil || n != 2 {
		t.Fatalf("n=%d err=%v", n, err)
	}
	var outbox, images int64
	db.Model(&ObjectDeletionEntity{}).Count(&outbox)
	db.Model(&Entity{}).Count(&images)
	if outbox != 4 || images != 1 {
		t.Fatalf("outbox=%d images=%d", outbox, images)
	}
	if n, err := PurgeByRecipeIDsTx(db, nil); err != nil || n != 0 {
		t.Fatalf("empty purge n=%d err=%v", n, err)
	}
}

func TestIsUniqueViolation(t *testing.T) {
	db := newTestDB(t)
	tid, hid, rid := uuid.New(), uuid.New(), uuid.New()
	e := seedImage(t, db, tid, hid, rid)
	dup := e
	dup.Id = uuid.New()
	err := db.Create(&dup).Error
	if err == nil || !isUniqueViolation(err) {
		t.Fatalf("err = %v, want unique violation", err)
	}
	if isUniqueViolation(errors.New("other")) {
		t.Fatal("false positive")
	}
}
