package recipeclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/google/uuid"
)

func urlQueryEscape(s string) string {
	return url.QueryEscape(s)
}

type PlanIngredient struct {
	ID                uuid.UUID
	Name              string
	DisplayName       string
	Quantity          float64
	Unit              string
	CategoryName      string
	CategorySortOrder int
	ExtraQuantities   []QuantityUnit
}

type QuantityUnit struct {
	Quantity float64
	Unit     string
}

type ingredientResponse struct {
	Data []struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Name              string  `json:"name"`
			DisplayName       *string `json:"display_name"`
			Quantity          float64 `json:"quantity"`
			Unit              string  `json:"unit"`
			CategoryName      *string `json:"category_name"`
			CategorySortOrder *int    `json:"category_sort_order"`
			ExtraQuantities   []struct {
				Quantity float64 `json:"quantity"`
				Unit     string  `json:"unit"`
			} `json:"extra_quantities"`
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

func (c *Client) GetPlanIngredients(planID uuid.UUID, accessToken string) ([]PlanIngredient, error) {
	url := fmt.Sprintf("%s/api/v1/meals/plans/%s/ingredients", c.baseURL, planID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
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
		return nil, fmt.Errorf("recipe-service returned %d: %s", resp.StatusCode, string(body))
	}

	var ingResp ingredientResponse
	if err := json.NewDecoder(resp.Body).Decode(&ingResp); err != nil {
		return nil, err
	}

	var ingredients []PlanIngredient
	for _, d := range ingResp.Data {
		id, _ := uuid.Parse(d.ID)
		name := d.Attributes.Name
		if d.Attributes.DisplayName != nil && *d.Attributes.DisplayName != "" {
			name = *d.Attributes.DisplayName
		}

		quantity := formatQuantity(d.Attributes.Quantity, d.Attributes.Unit)
		pi := PlanIngredient{
			ID:       id,
			Name:     name,
			Quantity: d.Attributes.Quantity,
			Unit:     d.Attributes.Unit,
		}
		if d.Attributes.CategoryName != nil {
			pi.CategoryName = *d.Attributes.CategoryName
		}
		if d.Attributes.CategorySortOrder != nil {
			pi.CategorySortOrder = *d.Attributes.CategorySortOrder
		}
		_ = quantity // we format at import time

		for _, eq := range d.Attributes.ExtraQuantities {
			pi.ExtraQuantities = append(pi.ExtraQuantities, QuantityUnit{
				Quantity: eq.Quantity,
				Unit:     eq.Unit,
			})
		}

		ingredients = append(ingredients, pi)
	}
	return ingredients, nil
}

type IngredientLookup struct {
	CanonicalID  uuid.UUID
	Name         string
	DisplayName  string
	CategoryID   *uuid.UUID
}

type ingredientLookupResponse struct {
	Data struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Attributes struct {
			Name        string  `json:"name"`
			DisplayName string  `json:"display_name"`
			CategoryId  *string `json:"category_id"`
		} `json:"attributes"`
	} `json:"data"`
}

// LookupIngredient asks recipe-service to resolve a free-form ingredient name
// to a canonical ingredient (matching by name, alias, or normalized variants).
// Returns (lookup, true, nil) on match, (nil, false, nil) on a clean miss
// (HTTP 404), and (nil, false, err) for any other error.
func (c *Client) LookupIngredient(name, accessToken string) (*IngredientLookup, bool, error) {
	if strings.TrimSpace(name) == "" {
		return nil, false, nil
	}
	url := fmt.Sprintf("%s/api/v1/ingredients/lookup?name=%s", c.baseURL, urlQueryEscape(name))
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, false, err
	}
	req.AddCookie(&http.Cookie{Name: "access_token", Value: accessToken})

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, false, nil
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, false, fmt.Errorf("recipe-service returned %d: %s", resp.StatusCode, string(body))
	}

	var lookup ingredientLookupResponse
	if err := json.NewDecoder(resp.Body).Decode(&lookup); err != nil {
		return nil, false, err
	}

	id, err := uuid.Parse(lookup.Data.ID)
	if err != nil {
		return nil, false, fmt.Errorf("invalid canonical ingredient id %q: %w", lookup.Data.ID, err)
	}
	result := &IngredientLookup{
		CanonicalID: id,
		Name:        lookup.Data.Attributes.Name,
		DisplayName: lookup.Data.Attributes.DisplayName,
	}
	if lookup.Data.Attributes.CategoryId != nil && *lookup.Data.Attributes.CategoryId != "" {
		cid, err := uuid.Parse(*lookup.Data.Attributes.CategoryId)
		if err != nil {
			return nil, false, fmt.Errorf("invalid category id %q: %w", *lookup.Data.Attributes.CategoryId, err)
		}
		result.CategoryID = &cid
	}
	return result, true, nil
}

func formatQuantity(qty float64, unit string) string {
	if qty == 0 {
		return ""
	}
	qtyStr := strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", qty), "0"), ".")
	if unit == "" {
		return qtyStr
	}
	return qtyStr + " " + unit
}

func FormatQuantityString(qty float64, unit string) string {
	return formatQuantity(qty, unit)
}
