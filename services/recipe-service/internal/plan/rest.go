package plan

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/audit"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/export"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/planitem"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/planner"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipe"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// planProviderImpl implements planitem.PlanProvider to break the import cycle.
type planProviderImpl struct {
	db *gorm.DB
}

func (pp *planProviderImpl) GetPlan(id uuid.UUID) (planitem.PlanInfo, error) {
	var e Entity
	if err := pp.db.Where("id = ?", id).First(&e).Error; err != nil {
		return planitem.PlanInfo{}, err
	}
	return planitem.PlanInfo{
		ID:       e.Id,
		StartsOn: e.StartsOn,
		Locked:   e.Locked,
	}, nil
}

func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		rihCreate := server.RegisterInputHandler[CreateRequest](l)(si)
		rihUpdate := server.RegisterInputHandler[UpdateRequest](l)(si)
		rihDuplicate := server.RegisterInputHandler[DuplicateRequest](l)(si)
		rihAddItem := server.RegisterInputHandler[planitem.CreateItemRequest](l)(si)
		rihUpdateItem := server.RegisterInputHandler[planitem.UpdateItemRequest](l)(si)

		pp := &planProviderImpl{db: db}

		api.HandleFunc("/meals/plans", rihCreate("CreatePlan", createPlanHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/meals/plans", rh("ListPlans", listPlansHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/meals/plans/{planId}", rh("GetPlan", getPlanHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/meals/plans/{planId}", rihUpdate("UpdatePlan", updatePlanHandler(db))).Methods(http.MethodPatch)
		api.HandleFunc("/meals/plans/{planId}/lock", rh("LockPlan", lockPlanHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/meals/plans/{planId}/unlock", rh("UnlockPlan", unlockPlanHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/meals/plans/{planId}/duplicate", rihDuplicate("DuplicatePlan", duplicatePlanHandler(db))).Methods(http.MethodPost)

		// Plan item routes
		api.HandleFunc("/meals/plans/{planId}/items", rihAddItem("AddPlanItem", planitem.AddItemHandler(db, pp))).Methods(http.MethodPost)
		api.HandleFunc("/meals/plans/{planId}/items/{itemId}", rihUpdateItem("UpdatePlanItem", planitem.UpdateItemHandler(db, pp))).Methods(http.MethodPatch)
		api.HandleFunc("/meals/plans/{planId}/items/{itemId}", rh("RemovePlanItem", planitem.RemoveItemHandler(db, pp))).Methods(http.MethodDelete)

		// Export routes
		api.HandleFunc("/meals/plans/{planId}/export/markdown", rh("ExportMarkdown", exportMarkdownHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/meals/plans/{planId}/ingredients", rh("GetPlanIngredients", getIngredientsHandler(db))).Methods(http.MethodGet)
	}
}

func createPlanHandler(db *gorm.DB) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			startsOn, err := time.Parse("2006-01-02", input.StartsOn)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Validation Failed", "starts_on must be a valid date (YYYY-MM-DD)")
				return
			}

			proc := NewProcessor(d.Logger(), r.Context(), db)
			m, err := proc.Create(t.Id(), t.HouseholdId(), t.UserId(), CreateAttrs{
				StartsOn: startsOn,
				Name:     input.Name,
			})
			if err != nil {
				if errors.Is(err, ErrAlreadyExists) {
					server.WriteError(w, http.StatusConflict, "Conflict", "A plan already exists for this household and week")
					return
				}
				if errors.Is(err, ErrStartsOnRequired) {
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
					return
				}
				d.Logger().WithError(err).Error("Failed to create plan")
				server.WriteError(w, http.StatusInternalServerError, "Create Failed", "")
				return
			}

			rest := TransformDetail(m, []RestItemModel{})
			server.MarshalCreatedResponse[RestDetailModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func listPlansHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			filters := ListFilters{
				Page:     queryInt(r, "page[number]", 1),
				PageSize: queryInt(r, "page[size]", 20),
			}

			if startsOnStr := r.URL.Query().Get("starts_on"); startsOnStr != "" {
				startsOn, err := time.Parse("2006-01-02", startsOnStr)
				if err == nil {
					filters.StartsOn = &startsOn
				}
			}

			proc := NewProcessor(d.Logger(), r.Context(), db)
			models, total, err := proc.List(filters)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to list plans")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			itemProc := planitem.NewProcessor(d.Logger(), r.Context(), db)
			rest := make([]RestListModel, len(models))
			for i, m := range models {
				count, _ := itemProc.CountByPlanWeekID(m.Id())
				rest[i] = TransformList(m, count)
			}

			items := make([]jsonapi.MarshalIdentifier, len(rest))
			for i := range rest {
				items[i] = &rest[i]
			}
			result, err := jsonapi.MarshalWithURLs(items, c.ServerInformation())
			if err != nil {
				d.Logger().WithError(err).Error("Failed to marshal response")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			var resp map[string]interface{}
			json.Unmarshal(result, &resp)
			resp["meta"] = map[string]interface{}{
				"total":    total,
				"page":     filters.Page,
				"pageSize": filters.PageSize,
			}

			w.Header().Set("Content-Type", "application/vnd.api+json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
		}
	}
}

func getPlanHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("planId", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.Get(id)
				if err != nil {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Plan not found")
					return
				}

				items := buildItemsResponse(d.Logger(), r.Context(), db, m)
				rest := TransformDetail(m, items)
				server.MarshalResponse[RestDetailModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func updatePlanHandler(db *gorm.DB) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input UpdateRequest) http.HandlerFunc {
		return server.ParseID("planId", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.UpdateName(id, input.Name)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Plan not found")
						return
					}
					if errors.Is(err, ErrLocked) {
						server.WriteError(w, http.StatusConflict, "Conflict", "Plan is locked")
						return
					}
					d.Logger().WithError(err).Error("Failed to update plan")
					server.WriteError(w, http.StatusInternalServerError, "Update Failed", "")
					return
				}

				items := buildItemsResponse(d.Logger(), r.Context(), db, m)
				rest := TransformDetail(m, items)
				server.MarshalResponse[RestDetailModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func lockPlanHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("planId", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.Lock(id)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Plan not found")
						return
					}
					if errors.Is(err, ErrAlreadyLocked) {
						server.WriteError(w, http.StatusConflict, "Conflict", "Plan is already locked")
						return
					}
					d.Logger().WithError(err).Error("Failed to lock plan")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				items := buildItemsResponse(d.Logger(), r.Context(), db, m)
				rest := TransformDetail(m, items)
				server.MarshalResponse[RestDetailModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func unlockPlanHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("planId", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.Unlock(id)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Plan not found")
						return
					}
					if errors.Is(err, ErrNotLocked) {
						server.WriteError(w, http.StatusConflict, "Conflict", "Plan is not locked")
						return
					}
					d.Logger().WithError(err).Error("Failed to unlock plan")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				items := buildItemsResponse(d.Logger(), r.Context(), db, m)
				rest := TransformDetail(m, items)
				server.MarshalResponse[RestDetailModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func duplicatePlanHandler(db *gorm.DB) server.InputHandler[DuplicateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input DuplicateRequest) http.HandlerFunc {
		return server.ParseID("planId", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())

				targetStartsOn, err := time.Parse("2006-01-02", input.StartsOn)
				if err != nil {
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", "starts_on must be a valid date (YYYY-MM-DD)")
					return
				}

				proc := NewProcessor(d.Logger(), r.Context(), db)
				source, err := proc.Get(id)
				if err != nil {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Source plan not found")
					return
				}

				newPlan, err := proc.Duplicate(id, t.Id(), t.HouseholdId(), t.UserId(), targetStartsOn)
				if err != nil {
					if errors.Is(err, ErrAlreadyExists) {
						server.WriteError(w, http.StatusConflict, "Conflict", "A plan already exists for the target week")
						return
					}
					d.Logger().WithError(err).Error("Failed to duplicate plan")
					server.WriteError(w, http.StatusInternalServerError, "Duplicate Failed", "")
					return
				}

				// Copy items with day offset
				dayOffset := int(targetStartsOn.Sub(source.StartsOn()).Hours() / 24)
				itemProc := planitem.NewProcessor(d.Logger(), r.Context(), db)
				if err := itemProc.CopyItems(source.Id(), newPlan.Id(), dayOffset); err != nil {
					d.Logger().WithError(err).Error("Failed to copy plan items")
				}

				items := buildItemsResponse(d.Logger(), r.Context(), db, newPlan)
				rest := TransformDetail(newPlan, items)
				server.MarshalCreatedResponse[RestDetailModel](d.Logger())(w)(c.ServerInformation())(rest)
			}
		})
	}
}

// buildItemsResponse enriches plan items with recipe metadata.
func buildItemsResponse(l logrus.FieldLogger, ctx context.Context, db *gorm.DB, m Model) []RestItemModel {
	itemProc := planitem.NewProcessor(l, ctx, db)
	items, err := itemProc.GetByPlanWeekID(m.Id())
	if err != nil {
		l.WithError(err).Error("Failed to get plan items")
		return []RestItemModel{}
	}

	recipeProc := recipe.NewProcessor(l, ctx, db)
	plannerProc := planner.NewProcessor(l, ctx, db)

	restItems := make([]RestItemModel, len(items))
	for i, item := range items {
		ri := RestItemModel{
			Id:                item.Id(),
			Day:               item.Day().Format("2006-01-02"),
			Slot:              item.Slot(),
			RecipeID:          item.RecipeID(),
			ServingMultiplier: item.ServingMultiplier(),
			PlannedServings:   item.PlannedServings(),
			Notes:             item.Notes(),
			Position:          item.Position(),
		}

		// Enrich with recipe metadata
		rm, _, recipeErr := recipeProc.Get(item.RecipeID())
		if recipeErr != nil {
			ri.RecipeDeleted = true
			ri.RecipeTitle = "(deleted recipe)"
		} else {
			ri.RecipeTitle = rm.Title()
			ri.RecipeServings = rm.Servings()
			ri.RecipeDeleted = false
		}

		// Get classification from planner config
		if pc, err := plannerProc.GetByRecipeID(item.RecipeID()); err == nil {
			ri.RecipeClassification = pc.Classification()
			if ri.RecipeServings == nil && pc.ServingsYield() != nil {
				ri.RecipeServings = pc.ServingsYield()
			}
		}

		restItems[i] = ri
	}
	return restItems
}

func exportMarkdownHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("planId", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.Get(id)
				if err != nil {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Plan not found")
					return
				}

				exportProc := export.NewProcessor(d.Logger(), r.Context(), db)
				markdown := exportProc.GenerateMarkdown(export.PlanData{
					ID: m.Id(), Name: m.Name(), StartsOn: m.StartsOn(),
				})

				// Emit audit event
				t, ok := tenantctx.FromContext(r.Context())
				if ok {
					audit.Emit(d.Logger(), db.WithContext(r.Context()), t.Id(), "plan", m.Id(), "plan.exported", t.UserId(), map[string]interface{}{
						"format": "markdown",
					})
				}

				w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(markdown))
			}
		})
	}
}

func getIngredientsHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("planId", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.Get(id)
				if err != nil {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Plan not found")
					return
				}

				exportProc := export.NewProcessor(d.Logger(), r.Context(), db)
				consolidated := exportProc.ConsolidateIngredients(export.PlanData{
					ID: m.Id(), Name: m.Name(), StartsOn: m.StartsOn(),
				})

				rest := make([]export.RestIngredientModel, len(consolidated))
				for i, ci := range consolidated {
					rest[i] = export.TransformIngredient(ci)
				}

				server.MarshalSliceResponse[export.RestIngredientModel](d.Logger())(w)(c.ServerInformation())(rest)
			}
		})
	}
}

func queryInt(r *http.Request, key string, defaultVal int) int {
	s := r.URL.Query().Get(key)
	if s == "" {
		return defaultVal
	}
	v, err := strconv.Atoi(s)
	if err != nil || v < 1 {
		return defaultVal
	}
	return v
}
