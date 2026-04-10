package categoryclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type Category struct {
	ID        uuid.UUID
	Name      string
	SortOrder int
	CreatedAt time.Time
	UpdatedAt time.Time
}

type categoryResponse struct {
	Data []struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Name      string    `json:"name"`
			SortOrder int       `json:"sort_order"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
		} `json:"attributes"`
	} `json:"data"`
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{},
	}
}

func (c *Client) ListCategories(accessToken string, tenantID, householdID uuid.UUID) ([]Category, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/api/v1/categories", nil)
	if err != nil {
		return nil, err
	}
	req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})
	// Forward tenant context. The auth-service issues JWTs with nil
	// tenant/household claims and the auth middleware falls back to these
	// headers — without them, category-service resolves the request as
	// the nil tenant and serves an unrelated auto-seeded set of defaults.
	if tenantID != uuid.Nil {
		req.Header.Set("X-Tenant-ID", tenantID.String())
	}
	if householdID != uuid.Nil {
		req.Header.Set("X-Household-ID", householdID.String())
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("category-service returned %d: %s", resp.StatusCode, string(body))
	}

	var catResp categoryResponse
	if err := json.NewDecoder(resp.Body).Decode(&catResp); err != nil {
		return nil, err
	}

	categories := make([]Category, len(catResp.Data))
	for i, d := range catResp.Data {
		id, _ := uuid.Parse(d.ID)
		categories[i] = Category{
			ID:        id,
			Name:      d.Attributes.Name,
			SortOrder: d.Attributes.SortOrder,
			CreatedAt: d.Attributes.CreatedAt,
			UpdatedAt: d.Attributes.UpdatedAt,
		}
	}
	return categories, nil
}

func (c *Client) GetCategoryByID(id uuid.UUID, accessToken string, tenantID, householdID uuid.UUID) (*Category, error) {
	cats, err := c.ListCategories(accessToken, tenantID, householdID)
	if err != nil {
		return nil, err
	}
	for _, cat := range cats {
		if cat.ID == id {
			return &cat, nil
		}
	}
	return nil, fmt.Errorf("category %s not found", id)
}
