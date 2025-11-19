package meal

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/apps/svc-meals/ai"
	"github.com/jtumidanski/home-hub/apps/svc-meals/ingredient"
	"github.com/jtumidanski/home-hub/apps/svc-meals/user"
	"github.com/jtumidanski/home-hub/packages/shared-go/model/ops"
	"github.com/jtumidanski/home-hub/packages/shared-go/rest/server"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// InitializeRoutes registers all meal-related routes using api2go pattern
func InitializeRoutes(si jsonapi.ServerInformation, aiClient *ai.Client) func(db *gorm.DB) server.RouteInitializer {
	return func(db *gorm.DB) server.RouteInitializer {
		return func(router *mux.Router, l logrus.FieldLogger) {
			// Meal CRUD endpoints
			router.HandleFunc("/meals", server.RegisterHandler(l)(si)("get-meals", listMealsHandler(db, aiClient))).Methods(http.MethodGet)
			router.HandleFunc("/meals", server.RegisterInputHandler[CreateRequest](l)(si)("create-meal", createMealHandler(db, aiClient))).Methods(http.MethodPost)
			router.HandleFunc("/meals/{id}", server.RegisterHandler(l)(si)("get-meal", getMealHandler(db, aiClient))).Methods(http.MethodGet)
			router.HandleFunc("/meals/{id}", server.RegisterInputHandler[UpdateRequest](l)(si)("update-meal", updateMealHandler(db, aiClient))).Methods(http.MethodPatch)
			router.HandleFunc("/meals/{id}", server.RegisterHandler(l)(si)("delete-meal", deleteMealHandler(db, aiClient))).Methods(http.MethodDelete)

			// Ingredient endpoints
			router.HandleFunc("/meals/{id}/ingredients", server.RegisterHandler(l)(si)("get-meal-ingredients", getIngredientsHandler(db, aiClient))).Methods(http.MethodGet)
		}
	}
}

// ParseId parses the mealId consistently from the request, and provide it to the next handler
func ParseId(l logrus.FieldLogger, next func(mealId uuid.UUID) http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id, err := uuid.Parse(vars["id"])
		if err != nil {
			l.WithError(err).Errorf("Unable to properly parse id from path.")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		next(id)(w, r)
	}
}

// listMealsHandler handles GET /meals
func listMealsHandler(db *gorm.DB, aiClient *ai.Client) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get user info from context
			userInfo, ok := user.UserInfoFromContext(r.Context())
			if !ok {
				d.Logger().Error("User not authenticated")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if userInfo.HouseholdID == uuid.Nil {
				d.Logger().Error("User is not associated with a household")
				w.WriteHeader(http.StatusForbidden)
				return
			}

			// List meals
			meals, err := ListByHousehold(r.Context(), db, userInfo.HouseholdID)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to list meals")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Transform to REST models
			restModels, err := TransformSlice(meals)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to transform meals")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			server.MarshalResponse[[]RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(restModels)
		}
	}
}

// getMealHandler handles GET /meals/{id}
func getMealHandler(db *gorm.DB, aiClient *ai.Client) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return ParseId(d.Logger(), func(mealID uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				// Get user info from context
				userInfo, ok := user.UserInfoFromContext(r.Context())
				if !ok {
					d.Logger().Error("User not authenticated")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				// Get meal
				meal, err := GetById(r.Context(), db, mealID)
				if err != nil {
					d.Logger().WithError(err).WithField("meal_id", mealID).Error("Failed to get meal")
					w.WriteHeader(http.StatusNotFound)
					return
				}

				// Verify user has access (same household)
				if meal.HouseholdId() != userInfo.HouseholdID {
					d.Logger().Error("Access denied")
					w.WriteHeader(http.StatusForbidden)
					return
				}

				// Transform to REST model
				res, err := ops.Map(Transform)(ops.FixedProvider(meal))()
				if err != nil {
					d.Logger().WithError(err).Error("Creating REST model.")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(res)
			}
		})
	}
}

// createMealHandler handles POST /meals
func createMealHandler(db *gorm.DB, aiClient *ai.Client) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, req CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get user info from context
			userInfo, ok := user.UserInfoFromContext(r.Context())
			if !ok {
				d.Logger().Error("User not authenticated")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			if userInfo.HouseholdID == uuid.Nil {
				d.Logger().Error("User is not associated with a household")
				w.WriteHeader(http.StatusForbidden)
				return
			}

			// Validate request
			if req.Title == "" {
				d.Logger().Error("Title is required")
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			// Create meal model
			meal, err := Create(userInfo.HouseholdID, userInfo.UserID, req.Title, req.Description, req.IngredientText)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to create meal")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Save meal
			if err := Save(r.Context(), db, meal); err != nil {
				d.Logger().WithError(err).Error("Failed to save meal")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Parse ingredients using AI if ingredient text is provided
			if req.IngredientText != "" {
				// Call AI service to parse recipe (extracts ingredients from full recipe text)
				// Returns builders (not complete models) because meal ID is not known until meal is saved
				ingredientBuilders, err := aiClient.ParseRecipe(r.Context(), req.IngredientText)
				if err != nil {
					d.Logger().WithError(err).Warn("Failed to parse recipe with AI, saving meal without parsed ingredients")
				} else if len(ingredientBuilders) > 0 {
					// Build complete ingredient models with meal ID
					ingredients := make([]ingredient.Model, 0, len(ingredientBuilders))
					for i, builder := range ingredientBuilders {
						ing, err := builder.ForMeal(meal.Id()).Build()
						if err != nil {
							d.Logger().WithError(err).WithField("index", i).Error("Failed to build ingredient from AI response")
							continue // Skip invalid ingredients instead of using zero values
						}
						ingredients = append(ingredients, ing)
					}

					// Save ingredients
					if len(ingredients) > 0 {
						if err := ingredient.SaveBatch(r.Context(), db, ingredients); err != nil {
							d.Logger().WithError(err).Error("Failed to save ingredients")
							// Don't fail the whole request - meal is already saved
						}
					}
				}
			}

			// Transform to REST model
			res, err := ops.Map(Transform)(ops.FixedProvider(meal))()
			if err != nil {
				d.Logger().WithError(err).Error("Creating REST model.")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(res)
		}
	}
}

// updateMealHandler handles PATCH /meals/{id}
func updateMealHandler(db *gorm.DB, aiClient *ai.Client) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, req UpdateRequest) http.HandlerFunc {
		return ParseId(d.Logger(), func(mealID uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				// Get user info from context
				userInfo, ok := user.UserInfoFromContext(r.Context())
				if !ok {
					d.Logger().Error("User not authenticated")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				// Get existing meal
				meal, err := GetById(r.Context(), db, mealID)
				if err != nil {
					d.Logger().WithError(err).Error("Meal not found")
					w.WriteHeader(http.StatusNotFound)
					return
				}

				// Verify user has access
				if meal.HouseholdId() != userInfo.HouseholdID {
					d.Logger().Error("Access denied")
					w.WriteHeader(http.StatusForbidden)
					return
				}

				// Update meal
				updated := Update(meal, req.Title, req.Description)

				// Save updated meal
				if err := Save(r.Context(), db, updated); err != nil {
					d.Logger().WithError(err).Error("Failed to update meal")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				// Transform to REST model
				res, err := ops.Map(Transform)(ops.FixedProvider(updated))()
				if err != nil {
					d.Logger().WithError(err).Error("Creating REST model.")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(res)
			}
		})
	}
}

// deleteMealHandler handles DELETE /meals/{id}
func deleteMealHandler(db *gorm.DB, aiClient *ai.Client) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return ParseId(d.Logger(), func(mealID uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				// Get user info from context
				userInfo, ok := user.UserInfoFromContext(r.Context())
				if !ok {
					d.Logger().Error("User not authenticated")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				// Get meal to verify access
				meal, err := GetById(r.Context(), db, mealID)
				if err != nil {
					d.Logger().WithError(err).Error("Meal not found")
					w.WriteHeader(http.StatusNotFound)
					return
				}

				// Verify user has access
				if meal.HouseholdId() != userInfo.HouseholdID {
					d.Logger().Error("Access denied")
					w.WriteHeader(http.StatusForbidden)
					return
				}

				// Delete ingredients first (cascade delete)
				if err := ingredient.DeleteByMealId(r.Context(), db, mealID); err != nil {
					d.Logger().WithError(err).Error("Failed to delete ingredients")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				// Delete meal
				if err := DeleteById(r.Context(), db, mealID); err != nil {
					d.Logger().WithError(err).Error("Failed to delete meal")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}

// getIngredientsHandler handles GET /meals/{id}/ingredients
func getIngredientsHandler(db *gorm.DB, aiClient *ai.Client) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return ParseId(d.Logger(), func(mealID uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				// Get user info from context
				userInfo, ok := user.UserInfoFromContext(r.Context())
				if !ok {
					d.Logger().Error("User not authenticated")
					w.WriteHeader(http.StatusUnauthorized)
					return
				}

				// Get meal to verify it exists and user has access
				meal, err := GetById(r.Context(), db, mealID)
				if err != nil {
					d.Logger().WithError(err).WithField("meal_id", mealID).Error("Failed to get meal")
					w.WriteHeader(http.StatusNotFound)
					return
				}

				// Verify user has access (same household)
				if meal.HouseholdId() != userInfo.HouseholdID {
					d.Logger().Error("Access denied")
					w.WriteHeader(http.StatusForbidden)
					return
				}

				// Get ingredients for this meal
				ingredients, err := ingredient.GetByMealId(r.Context(), db, mealID)
				if err != nil {
					d.Logger().WithError(err).WithField("meal_id", mealID).Error("Failed to get ingredients")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				// Transform to REST models
				restModels, err := ingredient.TransformSlice(ingredients)
				if err != nil {
					d.Logger().WithError(err).Error("Failed to transform ingredients")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				server.MarshalResponse[[]ingredient.RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(restModels)
			}
		})
	}
}

// ingredientFromModel creates a builder from an ingredient model (for updating with ForMeal)
func ingredientFromModel(m ingredient.Model) *ingredient.Builder {
	return ingredient.New().
		WithId(m.Id()).
		ForMeal(m.MealId()).
		WithRawLine(m.RawLine()).
		WithQuantity(m.Quantity()).
		WithQuantityRaw(m.QuantityRaw()).
		WithUnit(m.Unit()).
		WithUnitRaw(m.UnitRaw()).
		WithIngredient(m.Ingredient()).
		WithPreparation(m.Preparation()).
		WithNotes(m.Notes()).
		WithConfidence(m.Confidence()).
		WithTimestamps(m.CreatedAt(), m.UpdatedAt())
}
