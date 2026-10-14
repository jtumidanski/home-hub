package recipe

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus/hooks/test"
	"gorm.io/gorm"

	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipe/cooklang"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipeimage"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
)

func seedRecipeImage(t *testing.T, db *gorm.DB, tenantID, householdID, recipeID uuid.UUID) uuid.UUID {
	t.Helper()
	id := uuid.New()
	now := time.Now().UTC()
	e := recipeimage.Entity{
		Id: id, TenantId: tenantID, HouseholdId: householdID, RecipeId: recipeID,
		FullKey: "f/" + id.String(), ThumbKey: "t/" + id.String(), ContentType: "image/jpeg",
		Width: 1600, Height: 1067, FullBytes: 1, ThumbBytes: 1, CreatedAt: now, UpdatedAt: now,
	}
	if err := db.Create(&e).Error; err != nil {
		t.Fatal(err)
	}
	return id
}

func TestBuildListEnrichmentsImageOneQuery(t *testing.T) {
	db := setupTestDB(t)
	tid, hid := uuid.New(), uuid.New()
	ctx := tenantctx.WithContext(context.Background(), tenantctx.New(tid, hid, uuid.New()))
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, ctx, db)

	var models []Model
	for i := 0; i < 5; i++ {
		m, _, err := p.Create(tid, hid, CreateAttrs{Title: "R", Source: "Boil."})
		if err != nil {
			t.Fatal(err)
		}
		models = append(models, m)
	}
	imageID := seedRecipeImage(t, db, tid, hid, models[0].Id())

	imageQueries := 0
	if err := db.Callback().Query().After("gorm:query").Register("test:count_recipe_images", func(tx *gorm.DB) {
		if strings.Contains(tx.Statement.SQL.String(), "recipe_images") {
			imageQueries++
		}
	}); err != nil {
		t.Fatal(err)
	}

	enrichments := p.BuildListEnrichments(models)

	if imageQueries != 1 {
		t.Fatalf("recipe_images queries = %d, want 1", imageQueries)
	}
	img := enrichments[0].Image
	if img == nil || img.Width != 1600 || img.Height != 1067 ||
		!strings.HasSuffix(img.FullURL, "size=full&v="+imageID.String()) ||
		!strings.HasSuffix(img.ThumbURL, "size=thumb&v="+imageID.String()) {
		t.Fatalf("image = %+v", img)
	}
	for i := 1; i < len(enrichments); i++ {
		if enrichments[i].Image != nil {
			t.Fatalf("recipe %d unexpectedly has an image", i)
		}
	}
}

func TestImageEnrichmentOtherHousehold(t *testing.T) {
	db := setupTestDB(t)
	tid, hidA, hidB := uuid.New(), uuid.New(), uuid.New()
	l, _ := test.NewNullLogger()
	ctxA := tenantctx.WithContext(context.Background(), tenantctx.New(tid, hidA, uuid.New()))
	ctxB := tenantctx.WithContext(context.Background(), tenantctx.New(tid, hidB, uuid.New()))
	m, _, err := NewProcessor(l, ctxA, db).Create(tid, hidA, CreateAttrs{Title: "R", Source: "Boil."})
	if err != nil {
		t.Fatal(err)
	}
	seedRecipeImage(t, db, tid, hidA, m.Id())

	pB := NewProcessor(l, ctxB, db)
	if e := pB.BuildListEnrichments([]Model{m}); e[0].Image != nil {
		t.Fatal("household B must not see household A's image in list")
	}
	if d := pB.BuildDetailEnrichment(m, nil); d.Image != nil {
		t.Fatal("household B must not see household A's image in detail")
	}
}

func TestDetailEnrichmentImage(t *testing.T) {
	db := setupTestDB(t)
	tid, hid := uuid.New(), uuid.New()
	ctx := tenantctx.WithContext(context.Background(), tenantctx.New(tid, hid, uuid.New()))
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, ctx, db)
	m, _, err := p.Create(tid, hid, CreateAttrs{Title: "R", Source: "Boil."})
	if err != nil {
		t.Fatal(err)
	}
	if d := p.BuildDetailEnrichment(m, nil); d.Image != nil {
		t.Fatal("no image expected yet")
	}
	seedRecipeImage(t, db, tid, hid, m.Id())
	if d := p.BuildDetailEnrichment(m, nil); d.Image == nil || d.Image.Width != 1600 {
		t.Fatalf("detail image = %+v", d.Image)
	}
}

func TestImageSerializesAsNull(t *testing.T) {
	m, _ := NewBuilder().SetId(uuid.New()).SetTitle("T").SetSource("s").Build()
	list, _ := json.Marshal(Transform(m, ListEnrichment{}))
	if !strings.Contains(string(list), `"image":null`) {
		t.Fatalf("list json = %s", list)
	}
	detail, _ := json.Marshal(TransformDetail(m, cooklang.ParseResult{}, DetailEnrichment{}))
	if !strings.Contains(string(detail), `"image":null`) {
		t.Fatalf("detail json = %s", detail)
	}
}

func TestPatchWithImageAttributeDoesNotTouchImage(t *testing.T) {
	db := setupTestDB(t)
	tid, hid := uuid.New(), uuid.New()
	ctx := tenantctx.WithContext(context.Background(), tenantctx.New(tid, hid, uuid.New()))
	l, _ := test.NewNullLogger()
	p := NewProcessor(l, ctx, db)
	m, _, err := p.Create(tid, hid, CreateAttrs{Title: "R", Source: "Boil."})
	if err != nil {
		t.Fatal(err)
	}
	imageID := seedRecipeImage(t, db, tid, hid, m.Id())

	router := mux.NewRouter()
	api := router.PathPrefix("/api/v1").Subrouter()
	api.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rc := tenantctx.WithContext(r.Context(), tenantctx.New(tid, hid, uuid.New()))
			next.ServeHTTP(w, r.WithContext(rc))
		})
	})
	InitializeRoutes(db)(l, server.GetServerInformation(), api)

	body := `{"data":{"type":"recipes","id":"` + m.Id().String() + `","attributes":{"title":"New","image":{"thumbUrl":"/evil","fullUrl":"/evil","width":1,"height":1}}}}`
	req := httptest.NewRequest(http.MethodPatch, "/api/v1/recipes/"+m.Id().String(), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/vnd.api+json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", rec.Code, rec.Body.String())
	}

	var rows []recipeimage.Entity
	if err := db.Find(&rows).Error; err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].Id != imageID || rows[0].FullKey != "f/"+imageID.String() || rows[0].ThumbKey != "t/"+imageID.String() || rows[0].Width != 1600 {
		t.Fatalf("image row changed: %+v", rows)
	}
	var resp struct {
		Data struct {
			Attributes struct {
				Image *RestImageModel `json:"image"`
			} `json:"attributes"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}
	img := resp.Data.Attributes.Image
	if img == nil || !strings.HasSuffix(img.FullURL, "size=full&v="+imageID.String()) {
		t.Fatalf("response did not carry the stored image: %s", rec.Body.String())
	}
}
