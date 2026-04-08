package categoryclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
)

type Category struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	SortOrder int       `json:"sort_order"`
}

type categoryResponse struct {
	Data []struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Name      string `json:"name"`
			SortOrder int    `json:"sort_order"`
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

func (c *Client) GetCategory(categoryID uuid.UUID, accessToken string) (*Category, error) {
	categories, err := c.ListCategories(accessToken)
	if err != nil {
		return nil, err
	}
	for _, cat := range categories {
		if cat.ID == categoryID {
			return &cat, nil
		}
	}
	return nil, fmt.Errorf("category %s not found", categoryID)
}

func (c *Client) ListCategories(accessToken string) ([]Category, error) {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/api/v1/categories", nil)
	if err != nil {
		return nil, err
	}
	req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})

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
		}
	}
	return categories, nil
}
