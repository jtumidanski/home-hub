package geocoding

import (
	"fmt"

	"github.com/jtumidanski/home-hub/services/weather-service/internal/openmeteo"
)

type RestModel struct {
	Id        string  `json:"-"`
	Name      string  `json:"name"`
	Country   string  `json:"country"`
	Admin1    string  `json:"admin1"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func (r RestModel) GetName() string { return "geocoding-results" }
func (r RestModel) GetID() string   { return r.Id }
func (r *RestModel) SetID(id string) error {
	r.Id = id
	return nil
}

func Transform(g openmeteo.GeocodingResult) (RestModel, error) {
	return RestModel{
		Id:        fmt.Sprintf("%d", g.ID),
		Name:      g.Name,
		Country:   g.Country,
		Admin1:    g.Admin1,
		Latitude:  g.Latitude,
		Longitude: g.Longitude,
	}, nil
}

func TransformSlice(results []openmeteo.GeocodingResult) ([]RestModel, error) {
	out := make([]RestModel, len(results))
	for i, r := range results {
		rm, err := Transform(r)
		if err != nil {
			return nil, err
		}
		out[i] = rm
	}
	return out, nil
}
