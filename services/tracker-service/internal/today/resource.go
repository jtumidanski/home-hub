package today

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/tracker-service/internal/entry"
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
		api.HandleFunc("/trackers/today", rh("GetTodayTrackers", todayHandler(db))).Methods(http.MethodGet)
	}
}

func todayHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			now := time.Now().UTC().Truncate(24 * time.Hour)
			dow := int(now.Weekday())

			items, err := trackingitem.GetAllByUser(t.UserId())(db.WithContext(r.Context()))()
			if err != nil {
				d.Logger().WithError(err).Error("Failed to list trackers for today")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			type todayItem struct {
				Id          uuid.UUID       `json:"id"`
				Type        string          `json:"type"`
				Name        string          `json:"name"`
				ScaleType   string          `json:"scale_type"`
				ScaleConfig json.RawMessage `json:"scale_config"`
				Color       string          `json:"color"`
				SortOrder   int             `json:"sort_order"`
			}

			type todayEntry struct {
				Id             uuid.UUID       `json:"id"`
				Type           string          `json:"type"`
				TrackingItemId uuid.UUID       `json:"tracking_item_id"`
				Date           string          `json:"date"`
				Value          json.RawMessage `json:"value"`
				Skipped        bool            `json:"skipped"`
				Note           *string         `json:"note,omitempty"`
				Scheduled      bool            `json:"scheduled"`
			}

			var scheduledItems []todayItem
			var scheduledItemIDs []uuid.UUID

			for _, e := range items {
				m, err := trackingitem.Make(e)
				if err != nil {
					continue
				}

				snap, err := schedule.GetEffectiveSchedule(m.Id(), now)(db.WithContext(r.Context()))()
				if err != nil {
					continue
				}
				sm, err := schedule.Make(snap)
				if err != nil {
					continue
				}

				sched := sm.Schedule()
				isScheduled := false
				if len(sched) == 0 {
					isScheduled = true
				} else {
					for _, d := range sched {
						if d == dow {
							isScheduled = true
							break
						}
					}
				}

				if isScheduled {
					scheduledItems = append(scheduledItems, todayItem{
						Id:          m.Id(),
						Type:        "trackers",
						Name:        m.Name(),
						ScaleType:   m.ScaleType(),
						ScaleConfig: m.ScaleConfig(),
						Color:       m.Color(),
						SortOrder:   m.SortOrder(),
					})
					scheduledItemIDs = append(scheduledItemIDs, m.Id())
				}
			}

			var todayEntries []todayEntry
			for _, itemID := range scheduledItemIDs {
				e, err := entry.GetByItemAndDate(itemID, now)(db.WithContext(r.Context()))()
				if err != nil {
					continue
				}
				em, err := entry.Make(e)
				if err != nil {
					continue
				}
				todayEntries = append(todayEntries, todayEntry{
					Id:             em.Id(),
					Type:           "tracker-entries",
					TrackingItemId: em.TrackingItemID(),
					Date:           em.Date().Format("2006-01-02"),
					Value:          em.Value(),
					Skipped:        em.Skipped(),
					Note:           em.Note(),
					Scheduled:      true,
				})
			}

			result := struct {
				Data struct {
					Type       string `json:"type"`
					Attributes struct {
						Date string `json:"date"`
					} `json:"attributes"`
					Relationships struct {
						Items struct {
							Data interface{} `json:"data"`
						} `json:"items"`
						Entries struct {
							Data interface{} `json:"data"`
						} `json:"entries"`
					} `json:"relationships"`
				} `json:"data"`
			}{}

			result.Data.Type = "tracker-today"
			result.Data.Attributes.Date = now.Format("2006-01-02")
			if len(scheduledItems) > 0 {
				result.Data.Relationships.Items.Data = scheduledItems
			} else {
				result.Data.Relationships.Items.Data = []todayItem{}
			}
			if len(todayEntries) > 0 {
				result.Data.Relationships.Entries.Data = todayEntries
			} else {
				result.Data.Relationships.Entries.Data = []todayEntry{}
			}

			b, _ := json.Marshal(result)
			w.Header().Set("Content-Type", "application/vnd.api+json")
			w.WriteHeader(http.StatusOK)
			w.Write(b)
		}
	}
}
