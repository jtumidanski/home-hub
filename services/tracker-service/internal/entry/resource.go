package entry

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/schedule"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/trackingitem"
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

func isScheduledForDate(db *gorm.DB, ctx interface{ Done() <-chan struct{} }, itemID uuid.UUID, date time.Time) bool {
	snap, err := schedule.GetEffectiveSchedule(itemID, date)(db)()
	if err != nil {
		return true
	}
	m, err := schedule.Make(snap)
	if err != nil {
		return true
	}
	sched := m.Schedule()
	if len(sched) == 0 {
		return true
	}
	dow := int(date.Weekday())
	for _, d := range sched {
		if d == dow {
			return true
		}
	}
	return false
}

func createOrUpdateHandler(db *gorm.DB) server.InputHandler[EntryRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input EntryRequest) http.HandlerFunc {
		return server.ParseID("id", func(itemID uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())
				dateStr := mux.Vars(r)["date"]

				item, err := trackingitem.GetByID(itemID)(db.WithContext(r.Context()))()
				if err != nil {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Tracking item not found")
					return
				}

				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, created, err := proc.CreateOrUpdate(t.Id(), t.UserId(), itemID, dateStr, input.Value, input.Note, item.ScaleType, item.ScaleConfig)
				if err != nil {
					if errors.Is(err, ErrFutureDate) || errors.Is(err, ErrDateRequired) ||
						errors.Is(err, ErrInvalidSentiment) || errors.Is(err, ErrInvalidNumeric) ||
						errors.Is(err, ErrInvalidRange) || errors.Is(err, ErrValueRequired) ||
						errors.Is(err, ErrNoteTooLong) {
						server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
						return
					}
					d.Logger().WithError(err).Error("Failed to create/update entry")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				scheduled := isScheduledForDate(db.WithContext(r.Context()), r.Context(), itemID, m.Date())
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
				t := tenantctx.MustFromContext(r.Context())
				dateStr := mux.Vars(r)["date"]

				_, err := trackingitem.GetByID(itemID)(db.WithContext(r.Context()))()
				if err != nil {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Tracking item not found")
					return
				}

				date, err := time.Parse("2006-01-02", dateStr)
				if err != nil {
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", "invalid date format")
					return
				}

				scheduled := isScheduledForDate(db.WithContext(r.Context()), r.Context(), itemID, date)

				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.Skip(t.Id(), t.UserId(), itemID, dateStr, scheduled)
				if err != nil {
					if errors.Is(err, ErrNotScheduled) || errors.Is(err, ErrFutureDate) {
						server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
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
			models, err := proc.ListByMonth(t.UserId(), monthStr)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
				return
			}

			rest := make([]*RestModel, len(models))
			for i, m := range models {
				scheduled := isScheduledForDate(db.WithContext(r.Context()), r.Context(), m.TrackingItemID(), m.Date())
				rm := Transform(m, scheduled)
				rest[i] = &rm
			}

			server.MarshalSliceResponse[*RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}
