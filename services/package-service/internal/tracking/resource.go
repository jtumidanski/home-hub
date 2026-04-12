package tracking

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/package-service/internal/carrier"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitializeRoutes(db *gorm.DB, maxActive int, carriers *carrier.Registry) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		rihCreate := server.RegisterInputHandler[CreateRequest](l)(si)
		rihUpdate := server.RegisterInputHandler[UpdateRequest](l)(si)

		api.HandleFunc("/packages/summary", rh("PackageSummary", summaryHandler(db, maxActive, carriers))).Methods(http.MethodGet)
		api.HandleFunc("/packages/carriers/detect", rh("DetectCarrier", detectCarrierHandler(db, maxActive, carriers))).Methods(http.MethodGet)
		api.HandleFunc("/packages", rh("ListPackages", listHandler(db, maxActive, carriers))).Methods(http.MethodGet)
		api.HandleFunc("/packages", rihCreate("CreatePackage", createHandler(db, maxActive, carriers))).Methods(http.MethodPost)
		api.HandleFunc("/packages/{id}", rh("GetPackage", getHandler(db, maxActive, carriers))).Methods(http.MethodGet)
		api.HandleFunc("/packages/{id}", rihUpdate("UpdatePackage", updateHandler(db, maxActive, carriers))).Methods(http.MethodPatch)
		api.HandleFunc("/packages/{id}", rh("DeletePackage", deleteHandler(db, maxActive, carriers))).Methods(http.MethodDelete)
		api.HandleFunc("/packages/{id}/archive", rh("ArchivePackage", archiveHandler(db, maxActive, carriers))).Methods(http.MethodPost)
		api.HandleFunc("/packages/{id}/unarchive", rh("UnarchivePackage", unarchiveHandler(db, maxActive, carriers))).Methods(http.MethodPost)
		api.HandleFunc("/packages/{id}/refresh", rh("RefreshPackage", refreshHandler(db, maxActive, carriers))).Methods(http.MethodPost)
	}
}

func newProc(l logrus.FieldLogger, r *http.Request, db *gorm.DB, maxActive int, carriers *carrier.Registry) *Processor {
	return NewProcessor(l, r.Context(), db, maxActive, carriers)
}

func createHandler(db *gorm.DB, maxActive int, carriers *carrier.Registry) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			proc := newProc(d.Logger(), r, db, maxActive, carriers)
			m, err := proc.Create(t.Id(), t.HouseholdId(), t.UserId(), CreateAttrs{
				TrackingNumber: input.TrackingNumber,
				Carrier:        input.Carrier,
				Label:          input.Label,
				Notes:          input.Notes,
				Private:        input.Private,
			})
			if err != nil {
				if errors.Is(err, ErrTrackingNumberRequired) || errors.Is(err, ErrCarrierRequired) || errors.Is(err, ErrInvalidCarrier) {
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
					return
				}
				if errors.Is(err, ErrDuplicate) {
					server.WriteError(w, http.StatusConflict, "Conflict", err.Error())
					return
				}
				if errors.Is(err, ErrHouseholdLimit) {
					server.WriteError(w, http.StatusUnprocessableEntity, "Limit Reached", err.Error())
					return
				}
				d.Logger().WithError(err).Error("failed to create package")
				server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
				return
			}

			rest, err := TransformWithPrivacy(m, t.UserId())
			if err != nil {
				d.Logger().WithError(err).Error("transforming package")
				server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
				return
			}
			server.MarshalCreatedResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func listHandler(db *gorm.DB, maxActive int, carriers *carrier.Registry) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			includeArchived := r.URL.Query().Get("filter[archived]") == "true"
			hasETA := r.URL.Query().Get("filter[hasEta]") == "true"
			sortField := r.URL.Query().Get("sort")

			var filterStatuses []string
			if s := r.URL.Query().Get("filter[status]"); s != "" {
				filterStatuses = strings.Split(s, ",")
			}

			proc := newProc(d.Logger(), r, db, maxActive, carriers)
			models, err := proc.List(t.HouseholdId(), includeArchived, filterStatuses, hasETA, sortField)
			if err != nil {
				d.Logger().WithError(err).Error("failed to list packages")
				server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
				return
			}

			rest, err := TransformSliceWithPrivacy(models, t.UserId())
			if err != nil {
				d.Logger().WithError(err).Error("transforming packages")
				server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
				return
			}
			server.MarshalSliceResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func getHandler(db *gorm.DB, maxActive int, carriers *carrier.Registry) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())

				proc := newProc(d.Logger(), r, db, maxActive, carriers)
				m, err := proc.Get(id)
				if err != nil {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Package not found")
					return
				}

				if m.IsPrivate() && m.UserID() != t.UserId() {
					server.WriteError(w, http.StatusForbidden, "Forbidden", "This package is private")
					return
				}

				eventModels, err := proc.GetTrackingEvents(id)
				if err != nil {
					d.Logger().WithError(err).Error("failed to get tracking events")
					eventModels = nil
				}

				events := TransformTrackingEventSlice(eventModels)

				rest, err := TransformDetail(m, events, t.UserId())
				if err != nil {
					d.Logger().WithError(err).Error("transforming package detail")
					server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
					return
				}
				server.MarshalResponse[RestDetailModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func updateHandler(db *gorm.DB, maxActive int, carriers *carrier.Registry) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input UpdateRequest) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())

				proc := newProc(d.Logger(), r, db, maxActive, carriers)
				m, err := proc.Update(id, t.UserId(), UpdateAttrs{
					Label:   input.Label,
					Notes:   input.Notes,
					Carrier: input.Carrier,
					Private: input.Private,
				})
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Package not found")
						return
					}
					if errors.Is(err, ErrNotOwner) {
						server.WriteError(w, http.StatusForbidden, "Forbidden", err.Error())
						return
					}
					if errors.Is(err, ErrInvalidCarrier) {
						server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
						return
					}
					d.Logger().WithError(err).Error("failed to update package")
					server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
					return
				}

				rest, err := TransformWithPrivacy(m, t.UserId())
				if err != nil {
					d.Logger().WithError(err).Error("transforming package")
					server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
					return
				}
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func deleteHandler(db *gorm.DB, maxActive int, carriers *carrier.Registry) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())

				proc := newProc(d.Logger(), r, db, maxActive, carriers)
				if err := proc.Delete(id, t.UserId()); err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Package not found")
						return
					}
					if errors.Is(err, ErrNotOwner) {
						server.WriteError(w, http.StatusForbidden, "Forbidden", err.Error())
						return
					}
					d.Logger().WithError(err).Error("failed to delete package")
					server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}

func archiveHandler(db *gorm.DB, maxActive int, carriers *carrier.Registry) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())

				proc := newProc(d.Logger(), r, db, maxActive, carriers)
				m, err := proc.Archive(id)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Package not found")
						return
					}
					d.Logger().WithError(err).Error("failed to archive package")
					server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
					return
				}

				rest, err := TransformWithPrivacy(m, t.UserId())
				if err != nil {
					d.Logger().WithError(err).Error("transforming package")
					server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
					return
				}
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func unarchiveHandler(db *gorm.DB, maxActive int, carriers *carrier.Registry) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())

				proc := newProc(d.Logger(), r, db, maxActive, carriers)
				m, err := proc.Unarchive(id)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Package not found")
						return
					}
					if errors.Is(err, ErrNotArchived) {
						server.WriteError(w, http.StatusBadRequest, "Bad Request", err.Error())
						return
					}
					d.Logger().WithError(err).Error("failed to unarchive package")
					server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
					return
				}

				rest, err := TransformWithPrivacy(m, t.UserId())
				if err != nil {
					d.Logger().WithError(err).Error("transforming package")
					server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
					return
				}
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func summaryHandler(db *gorm.DB, maxActive int, carriers *carrier.Registry) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			proc := newProc(d.Logger(), r, db, maxActive, carriers)
			result, err := proc.Summary(t.HouseholdId())
			if err != nil {
				d.Logger().WithError(err).Error("failed to get package summary")
				server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
				return
			}

			rest := RestSummaryModel{
				ArrivingTodayCount: result.ArrivingTodayCount,
				InTransitCount:     result.InTransitCount,
				ExceptionCount:     result.ExceptionCount,
			}
			server.MarshalResponse[RestSummaryModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
		}
	}
}

func detectCarrierHandler(db *gorm.DB, maxActive int, carriers *carrier.Registry) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			tn := r.URL.Query().Get("trackingNumber")
			if tn == "" {
				server.WriteError(w, http.StatusBadRequest, "Missing Parameter", "trackingNumber query parameter is required")
				return
			}

			proc := newProc(d.Logger(), r, db, maxActive, carriers)
			result := proc.DetectCarrier(tn)
			rest := TransformDetection(result)
			server.MarshalResponse[carrier.RestDetectionModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
		}
	}
}

func refreshHandler(db *gorm.DB, maxActive int, carriers *carrier.Registry) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())

				proc := newProc(d.Logger(), r, db, maxActive, carriers)
				m, err := proc.Refresh(id)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Package not found")
						return
					}
					if errors.Is(err, ErrRefreshTooSoon) {
						server.WriteError(w, http.StatusTooManyRequests, "Too Many Requests", fmt.Sprintf("Package was recently refreshed. Try again in %d minutes.", int(refreshCooldown.Minutes())))
						return
					}
					d.Logger().WithError(err).Error("failed to refresh package")
					server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
					return
				}

				rest, err := TransformWithPrivacy(m, t.UserId())
				if err != nil {
					d.Logger().WithError(err).Error("transforming package")
					server.WriteError(w, http.StatusInternalServerError, "Internal Error", "")
					return
				}
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}
