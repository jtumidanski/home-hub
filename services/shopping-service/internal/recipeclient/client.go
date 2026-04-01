package recipeclient

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

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

func (c *Client) GetPlanIngredients(planID uuid.UUID, authHeader string) ([]PlanIngredient, error) {
	url := fmt.Sprintf("%s/api/v1/meals/plans/%s/ingredients", c.baseURL, planID)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", authHeader)

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
