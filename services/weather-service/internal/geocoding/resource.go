package geocoding

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/weather-service/internal/openmeteo"
	"github.com/jtumidanski/home-hub/shared/go/server"
	"github.com/sirupsen/logrus"
)

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

			proc := NewProcessor(d.Logger(), r.Context(), client)
			results, err := proc.Search(q)
			if err != nil {
				d.Logger().WithError(err).Error("Geocoding search failed")
				server.WriteError(w, http.StatusBadGateway, "Geocoding Unavailable", "Unable to search for places. Please try again later.")
				return
			}

			rest, err := TransformSlice(results)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to transform geocoding results")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			server.MarshalSliceResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}
