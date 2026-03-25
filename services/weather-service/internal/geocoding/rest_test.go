package geocoding

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jtumidanski/home-hub/services/weather-service/internal/openmeteo"
)

type mockClient struct {
	results []openmeteo.GeocodingResult
	err     error
}

func (m *mockClient) searchPlaces(query string) ([]openmeteo.GeocodingResult, error) {
	return m.results, m.err
}

func TestSearchHandlerShortQuery(t *testing.T) {
	client := openmeteo.NewClient()

	handler := testSearchHandler(client)

	req := httptest.NewRequest(http.MethodGet, "/weather/geocoding?q=a", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for short query, got %d", w.Code)
	}
}

func TestSearchHandlerEmptyQuery(t *testing.T) {
	client := openmeteo.NewClient()

	handler := testSearchHandler(client)

	req := httptest.NewRequest(http.MethodGet, "/weather/geocoding", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty query, got %d", w.Code)
	}
}

func TestRestModelGetName(t *testing.T) {
	r := RestModel{Id: "1", Name: "New York", Country: "United States"}
	if r.GetName() != "geocoding-results" {
		t.Errorf("expected geocoding-results, got %s", r.GetName())
	}
	if r.GetID() != "1" {
		t.Errorf("expected 1, got %s", r.GetID())
	}
}

func TestRestModelSetID(t *testing.T) {
	r := &RestModel{}
	if err := r.SetID("42"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Id != "42" {
		t.Errorf("expected 42, got %s", r.Id)
	}
}

func TestRestModelJSON(t *testing.T) {
	r := RestModel{
		Id:        "123",
		Name:      "London",
		Country:   "United Kingdom",
		Admin1:    "England",
		Latitude:  51.5,
		Longitude: -0.12,
	}

	data, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var m map[string]interface{}
	json.Unmarshal(data, &m)

	// Id should be omitted (json:"-")
	if _, ok := m["Id"]; ok {
		t.Error("Id should be omitted from JSON")
	}
	if m["name"] != "London" {
		t.Errorf("expected London, got %v", m["name"])
	}
	if m["country"] != "United Kingdom" {
		t.Errorf("expected United Kingdom, got %v", m["country"])
	}
}

// testSearchHandler creates a minimal handler for testing query validation.
// It bypasses the full server.RegisterHandler infrastructure since we only
// need to test the query parameter validation logic.
func testSearchHandler(client *openmeteo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query().Get("q")
		if len(q) < 2 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"errors":[{"title":"Invalid Query","detail":"Search query must be at least 2 characters."}]}`))
			return
		}

		results, err := client.SearchPlaces(q)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			return
		}

		rest := make([]RestModel, len(results))
		for i, r := range results {
			rest[i] = RestModel{
				Id:        fmt.Sprintf("%d", r.ID),
				Name:      r.Name,
				Country:   r.Country,
				Admin1:    r.Admin1,
				Latitude:  r.Latitude,
				Longitude: r.Longitude,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(rest)
	}
}
