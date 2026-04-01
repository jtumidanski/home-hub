package planitem

import (
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jtumidanski/home-hub/shared/go/server"
	"gorm.io/gorm"
)

// PlanProvider allows the plan item handlers to look up a plan without importing the plan package.
type PlanProvider interface {
	GetPlan(id uuid.UUID) (PlanInfo, error)
}

// PlanInfo is the minimal plan data needed by item handlers.
type PlanInfo struct {
	ID       uuid.UUID
	StartsOn time.Time
	Locked   bool
}

func AddItemHandler(db *gorm.DB, pp PlanProvider) server.InputHandler[CreateItemRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CreateItemRequest) http.HandlerFunc {
		return server.ParseID("planId", func(planID uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				p, err := pp.GetPlan(planID)
				if err != nil {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Plan not found")
					return
				}
				if p.Locked {
					server.WriteError(w, http.StatusConflict, "Conflict", "Plan is locked")
					return
				}

				day, err := time.Parse("2006-01-02", input.Day)
				if err != nil {
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", "day must be a valid date (YYYY-MM-DD)")
					return
				}

				recipeID, err := uuid.Parse(input.RecipeID)
				if err != nil {
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", "recipe_id must be a valid UUID")
					return
				}

				proc := NewProcessor(d.Logger(), r.Context(), db)

				// Validate recipe exists and is active
				if err := proc.ValidateRecipeExists(recipeID); err != nil {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Recipe not found or deleted")
					return
				}
				m, err := proc.AddItem(planID, p.StartsOn, AddAttrs{
					Day:               day,
					Slot:              input.Slot,
					RecipeID:          recipeID,
					ServingMultiplier: input.ServingMultiplier,
					PlannedServings:   input.PlannedServings,
					Notes:             input.Notes,
					Position:          input.Position,
				})
				if err != nil {
					if errors.Is(err, ErrDayOutOfRange) {
						server.WriteError(w, http.StatusBadRequest, "Validation Failed", "day must fall within the plan week")
						return
					}
					if errors.Is(err, ErrInvalidSlot) {
						server.WriteError(w, http.StatusBadRequest, "Validation Failed", "slot must be one of: breakfast, lunch, dinner, snack, side")
						return
					}
					d.Logger().WithError(err).Error("Failed to add plan item")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				rest := TransformItem(m)
				server.MarshalCreatedResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
			}
		})
	}
}

func UpdateItemHandler(db *gorm.DB, pp PlanProvider) server.InputHandler[UpdateItemRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input UpdateItemRequest) http.HandlerFunc {
		return server.ParseID("planId", func(planID uuid.UUID) http.HandlerFunc {
			return server.ParseID("itemId", func(itemID uuid.UUID) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					p, err := pp.GetPlan(planID)
					if err != nil {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Plan not found")
						return
					}
					if p.Locked {
						server.WriteError(w, http.StatusConflict, "Conflict", "Plan is locked")
						return
					}

					attrs := UpdateAttrs{}
					if input.Day != "" {
						day, err := time.Parse("2006-01-02", input.Day)
						if err != nil {
							server.WriteError(w, http.StatusBadRequest, "Validation Failed", "day must be a valid date (YYYY-MM-DD)")
							return
						}
						attrs.Day = &day
					}
					if input.Slot != "" {
						attrs.Slot = &input.Slot
					}
					if input.ServingMultiplier != nil {
						attrs.ServingMultiplier = &input.ServingMultiplier
					}
					if input.PlannedServings != nil {
						attrs.PlannedServings = &input.PlannedServings
					}
					if input.Notes != nil {
						attrs.Notes = &input.Notes
					}
					if input.Position != nil {
						attrs.Position = input.Position
					}

					proc := NewProcessor(d.Logger(), r.Context(), db)
					m, err := proc.UpdateItem(itemID, p.StartsOn, attrs)
					if err != nil {
						if errors.Is(err, ErrNotFound) {
							server.WriteError(w, http.StatusNotFound, "Not Found", "Plan item not found")
							return
						}
						if errors.Is(err, ErrDayOutOfRange) {
							server.WriteError(w, http.StatusBadRequest, "Validation Failed", "day must fall within the plan week")
							return
						}
						if errors.Is(err, ErrInvalidSlot) {
							server.WriteError(w, http.StatusBadRequest, "Validation Failed", "slot must be one of: breakfast, lunch, dinner, snack, side")
							return
						}
						d.Logger().WithError(err).Error("Failed to update plan item")
						server.WriteError(w, http.StatusInternalServerError, "Error", "")
						return
					}

					rest := TransformItem(m)
					server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
				}
			})
		})
	}
}

func RemoveItemHandler(db *gorm.DB, pp PlanProvider) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("planId", func(planID uuid.UUID) http.HandlerFunc {
			return server.ParseID("itemId", func(itemID uuid.UUID) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					p, err := pp.GetPlan(planID)
					if err != nil {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Plan not found")
						return
					}
					if p.Locked {
						server.WriteError(w, http.StatusConflict, "Conflict", "Plan is locked")
						return
					}

					proc := NewProcessor(d.Logger(), r.Context(), db)
					if err := proc.RemoveItem(itemID, planID); err != nil {
						if errors.Is(err, ErrNotFound) {
							server.WriteError(w, http.StatusNotFound, "Not Found", "Plan item not found")
							return
						}
						d.Logger().WithError(err).Error("Failed to remove plan item")
						server.WriteError(w, http.StatusInternalServerError, "Error", "")
						return
					}

					w.WriteHeader(http.StatusNoContent)
				}
			})
		})
	}
}

