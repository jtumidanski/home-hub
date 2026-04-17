package entry

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	httpparams "github.com/jtumidanski/home-hub/shared/go/http"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		rihEntry := server.RegisterInputHandler[EntryRequest](l)(si)

		api.HandleFunc("/trackers/{id}/entries/{date}", rihEntry("CreateOrUpdateEntry", createOrUpdateHandler(db))).Methods(http.MethodPut)
		api.HandleFunc("/trackers/{id}/entries/{date}", rh("DeleteEntry", deleteEntryHandler(db))).Methods(http.MethodDelete)
		api.HandleFunc("/trackers/{id}/entries/{date}/skip", rh("SkipEntry", skipHandler(db))).Methods(http.MethodPut)
		api.HandleFunc("/trackers/{id}/entries/{date}/skip", rh("RemoveSkip", removeSkipHandler(db))).Methods(http.MethodDelete)
		api.HandleFunc("/trackers/entries", rh("ListEntriesByMonth", listByMonthHandler(db))).Methods(http.MethodGet)
	}
}

func writeValidationError(w http.ResponseWriter, err error) bool {
	switch {
	case errors.Is(err, ErrFutureDate),
		errors.Is(err, ErrDateRequired),
		errors.Is(err, ErrInvalidSentiment),
		errors.Is(err, ErrInvalidNumeric),
		errors.Is(err, ErrInvalidRange),
		errors.Is(err, ErrValueRequired),
		errors.Is(err, ErrNoteTooLong),
		errors.Is(err, ErrNotScheduled):
		server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
		return true
	case errors.Is(err, ErrItemNotFound):
		server.WriteError(w, http.StatusNotFound, "Not Found", err.Error())
		return true
	}
	return false
}

func createOrUpdateHandler(db *gorm.DB) server.InputHandler[EntryRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input EntryRequest) http.HandlerFunc {
		return server.ParseID("id", func(itemID uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				today, err := httpparams.ParseDateParam(r, "today")
				if err != nil {
					server.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
					return
				}

				t := tenantctx.MustFromContext(r.Context())
				dateStr := mux.Vars(r)["date"]

				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, created, scheduled, err := proc.CreateOrUpdate(t.Id(), t.UserId(), itemID, dateStr, today, input.Value, input.Note)
				if err != nil {
					if writeValidationError(w, err) {
						return
					}
					d.Logger().WithError(err).Error("Failed to create/update entry")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				rest := Transform(m, scheduled)
				if created {
					server.MarshalCreatedResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
				} else {
					server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
				}
			}
		})
	}
}

func deleteEntryHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(itemID uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				dateStr := mux.Vars(r)["date"]
				proc := NewProcessor(d.Logger(), r.Context(), db)

				if err := proc.Delete(itemID, dateStr); err != nil {
					d.Logger().WithError(err).Error("Failed to delete entry")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}

func skipHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(itemID uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				today, err := httpparams.ParseDateParam(r, "today")
				if err != nil {
					server.WriteError(w, http.StatusBadRequest, "Invalid request", err.Error())
					return
				}

				t := tenantctx.MustFromContext(r.Context())
				dateStr := mux.Vars(r)["date"]

				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.Skip(t.Id(), t.UserId(), itemID, dateStr, today)
				if err != nil {
					if writeValidationError(w, err) {
						return
					}
					d.Logger().WithError(err).Error("Failed to skip entry")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				rest := Transform(m, true)
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func removeSkipHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(itemID uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				dateStr := mux.Vars(r)["date"]
				proc := NewProcessor(d.Logger(), r.Context(), db)

				if err := proc.RemoveSkip(itemID, dateStr); err != nil {
					d.Logger().WithError(err).Error("Failed to remove skip")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}

func listByMonthHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			monthStr := r.URL.Query().Get("month")
			if monthStr == "" {
				server.WriteError(w, http.StatusBadRequest, "Validation Failed", "month query parameter is required (YYYY-MM)")
				return
			}

			proc := NewProcessor(d.Logger(), r.Context(), db)
			results, err := proc.ListByMonthWithScheduled(t.UserId(), monthStr)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
				return
			}

			rest := TransformSlice(results)
			server.MarshalSliceResponse[*RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}
