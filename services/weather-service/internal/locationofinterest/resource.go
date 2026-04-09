package locationofinterest

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitializeRoutes(db *gorm.DB, makeWarmer func(l logrus.FieldLogger, r *http.Request) CacheWarmer) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		rihCreate := server.RegisterInputHandler[CreateRequest](l)(si)
		rihUpdate := server.RegisterInputHandler[UpdateRequest](l)(si)

		api.HandleFunc("/locations-of-interest", rh("ListLocationsOfInterest", listHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/locations-of-interest", rihCreate("CreateLocationOfInterest", createHandler(db, makeWarmer))).Methods(http.MethodPost)
		api.HandleFunc("/locations-of-interest/{id}", rihUpdate("UpdateLocationOfInterest", updateHandler(db))).Methods(http.MethodPatch)
		api.HandleFunc("/locations-of-interest/{id}", rh("DeleteLocationOfInterest", deleteHandler(db))).Methods(http.MethodDelete)
	}
}

func listHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db, nil)

			models, err := proc.List(t.HouseholdId())
			if err != nil {
				d.Logger().WithError(err).Error("Failed to list locations of interest")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			transformed, err := TransformSlice(models)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to transform locations of interest")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			rest := make([]*RestModel, len(transformed))
			for i := range transformed {
				rest[i] = &transformed[i]
			}
			server.MarshalSliceResponse[*RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func createHandler(db *gorm.DB, makeWarmer func(l logrus.FieldLogger, r *http.Request) CacheWarmer) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			var warmer CacheWarmer
			if makeWarmer != nil {
				warmer = makeWarmer(d.Logger(), r)
			}
			proc := NewProcessor(d.Logger(), r.Context(), db, warmer)

			m, err := proc.Create(t.Id(), t.HouseholdId(), CreateInput{
				Label:     input.Label,
				PlaceName: input.PlaceName,
				Latitude:  input.Latitude,
				Longitude: input.Longitude,
			})
			if err != nil {
				if errors.Is(err, ErrCapReached) {
					server.WriteError(w, http.StatusConflict, "Cap Reached", err.Error())
					return
				}
				if errors.Is(err, ErrLabelTooLong) ||
					errors.Is(err, ErrPlaceNameRequired) ||
					errors.Is(err, ErrLatitudeOutOfRange) ||
					errors.Is(err, ErrLongitudeOutOfRange) {
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
					return
				}
				d.Logger().WithError(err).Error("Failed to create location of interest")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			rest, err := Transform(m)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to transform location of interest")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			server.MarshalCreatedResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func updateHandler(db *gorm.DB) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input UpdateRequest) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())
				proc := NewProcessor(d.Logger(), r.Context(), db, nil)

				m, err := proc.UpdateLabel(t.HouseholdId(), id, input.Label)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Location of interest not found")
						return
					}
					if errors.Is(err, ErrLabelTooLong) {
						server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
						return
					}
					d.Logger().WithError(err).Error("Failed to update location of interest")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				rest, err := Transform(m)
				if err != nil {
					d.Logger().WithError(err).Error("Failed to transform location of interest")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func deleteHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())
				proc := NewProcessor(d.Logger(), r.Context(), db, nil)

				if err := proc.Delete(t.HouseholdId(), id); err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Location of interest not found")
						return
					}
					d.Logger().WithError(err).Error("Failed to delete location of interest")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}
