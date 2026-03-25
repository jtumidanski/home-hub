package geocoding

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/weather-service/internal/openmeteo"
	"github.com/jtumidanski/home-hub/shared/go/server"
	"github.com/sirupsen/logrus"
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

func InitializeRoutes(client *openmeteo.Client) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		api.HandleFunc("/weather/geocoding", rh("SearchGeocode", searchHandler(client))).Methods(http.MethodGet)
	}
}

func searchHandler(client *openmeteo.Client) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			q := r.URL.Query().Get("q")
			if len(q) < 2 {
				server.WriteError(w, http.StatusBadRequest, "Invalid Query", "Search query must be at least 2 characters.")
				return
			}

			results, err := client.SearchPlaces(q)
			if err != nil {
				d.Logger().WithError(err).Error("Geocoding search failed")
				server.WriteError(w, http.StatusBadGateway, "Geocoding Unavailable", "Unable to search for places. Please try again later.")
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

			server.MarshalSliceResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}
