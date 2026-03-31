package ingredient

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		rihCreate := server.RegisterInputHandler[CreateRequest](l)(si)
		rihUpdate := server.RegisterInputHandler[UpdateRequest](l)(si)
		rihAlias := server.RegisterInputHandler[AddAliasRequest](l)(si)
		rihReassign := server.RegisterInputHandler[ReassignRequest](l)(si)
		rihBulk := server.RegisterInputHandler[BulkCategorizeRequest](l)(si)

		api.HandleFunc("/ingredients", rh("ListIngredients", listIngredientsHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/ingredients/bulk-categorize", rihBulk("BulkCategorize", bulkCategorizeHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/ingredients", rihCreate("CreateIngredient", createIngredientHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/ingredients/{id}", rh("GetIngredient", getIngredientHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/ingredients/{id}", rihUpdate("UpdateIngredient", updateIngredientHandler(db))).Methods(http.MethodPatch)
		api.HandleFunc("/ingredients/{id}", rh("DeleteIngredient", deleteIngredientHandler(db))).Methods(http.MethodDelete)
		api.HandleFunc("/ingredients/{id}/aliases", rihAlias("AddAlias", addAliasHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/ingredients/{id}/aliases/{aliasId}", rh("RemoveAlias", removeAliasHandler(db))).Methods(http.MethodDelete)
		api.HandleFunc("/ingredients/{id}/recipes", rh("IngredientRecipes", ingredientRecipesHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/ingredients/{id}/reassign", rihReassign("ReassignIngredient", reassignHandler(db))).Methods(http.MethodPost)
	}
}

func listIngredientsHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			query := r.URL.Query().Get("search")
			categoryFilter := r.URL.Query().Get("filter[category_id]")
			page := queryInt(r, "page[number]", 1)
			pageSize := queryInt(r, "page[size]", 20)

			proc := NewProcessor(d.Logger(), r.Context(), db)
			models, total, err := proc.SearchWithUsage(t.Id(), query, categoryFilter, page, pageSize)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to list ingredients")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			rest := make([]RestModel, len(models))
			for i, m := range models {
				rest[i] = TransformList(m, m.UsageCount())
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
				"page":     page,
				"pageSize": pageSize,
			}

			w.Header().Set("Content-Type", "application/vnd.api+json")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(resp)
		}
	}
}

func createIngredientHandler(db *gorm.DB) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			var categoryID *uuid.UUID
			if input.CategoryId != nil && *input.CategoryId != "" {
				cid, err := uuid.Parse(*input.CategoryId)
				if err != nil {
					server.WriteError(w, http.StatusUnprocessableEntity, "Validation Failed", "categoryId must be a valid UUID")
					return
				}
				categoryID = &cid
			}

			proc := NewProcessor(d.Logger(), r.Context(), db)
			m, err := proc.Create(t.Id(), input.Name, input.DisplayName, input.UnitFamily, categoryID)
			if err != nil {
				if errors.Is(err, ErrNameRequired) {
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
					return
				}
				if errors.Is(err, ErrInvalidUnitFamily) {
					server.WriteError(w, http.StatusUnprocessableEntity, "Validation Failed", err.Error())
					return
				}
				d.Logger().WithError(err).Error("Failed to create ingredient")
				server.WriteError(w, http.StatusInternalServerError, "Create Failed", "")
				return
			}

			rest := TransformDetail(m)
			server.MarshalCreatedResponse[RestDetailModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func getIngredientHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.Get(id)
				if err != nil {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Ingredient not found")
					return
				}
				rest := TransformDetail(m)
				server.MarshalResponse[RestDetailModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func updateIngredientHandler(db *gorm.DB) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input UpdateRequest) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				var name, displayName, unitFamily *string
				if input.Name != "" {
					name = &input.Name
				}
				if input.DisplayName != "" {
					displayName = &input.DisplayName
				}
				if input.UnitFamily != "" {
					unitFamily = &input.UnitFamily
				}

				var categoryOpt *UpdateCategoryOpt
				if input.CategoryId != nil {
					opt := UpdateCategoryOpt{Set: true}
					if *input.CategoryId != "" {
						cid, err := uuid.Parse(*input.CategoryId)
						if err != nil {
							server.WriteError(w, http.StatusUnprocessableEntity, "Validation Failed", "categoryId must be a valid UUID")
							return
						}
						opt.Value = &cid
					}
					categoryOpt = &opt
				}

				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.Update(id, name, displayName, unitFamily, categoryOpt)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Ingredient not found")
						return
					}
					if errors.Is(err, ErrInvalidUnitFamily) {
						server.WriteError(w, http.StatusUnprocessableEntity, "Validation Failed", err.Error())
						return
					}
					d.Logger().WithError(err).Error("Failed to update ingredient")
					server.WriteError(w, http.StatusInternalServerError, "Update Failed", "")
					return
				}

				rest := TransformDetail(m)
				server.MarshalResponse[RestDetailModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func deleteIngredientHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				proc := NewProcessor(d.Logger(), r.Context(), db)
				if err := proc.Delete(id); err != nil {
					if errors.Is(err, ErrHasReferences) {
						server.WriteError(w, http.StatusConflict, "Conflict", "Ingredient is still referenced by recipe ingredients. Use reassign endpoint first.")
						return
					}
					d.Logger().WithError(err).Error("Failed to delete ingredient")
					server.WriteError(w, http.StatusInternalServerError, "Delete Failed", "")
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}

func addAliasHandler(db *gorm.DB) server.InputHandler[AddAliasRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input AddAliasRequest) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())
				proc := NewProcessor(d.Logger(), r.Context(), db)
				_, err := proc.AddAlias(t.Id(), id, input.Name)
				if err != nil {
					if errors.Is(err, ErrAliasConflict) {
						server.WriteError(w, http.StatusConflict, "Conflict", err.Error())
						return
					}
					d.Logger().WithError(err).Error("Failed to add alias")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				// Return updated ingredient
				m, err := proc.Get(id)
				if err != nil {
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				rest := TransformDetail(m)
				server.MarshalCreatedResponse[RestDetailModel](d.Logger())(w)(c.ServerInformation())(rest)
			}
		})
	}
}

func removeAliasHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(_ uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				aliasIDStr := mux.Vars(r)["aliasId"]
				aliasID, err := uuid.Parse(aliasIDStr)
				if err != nil {
					server.WriteError(w, http.StatusBadRequest, "Invalid Request", "aliasId must be a valid UUID")
					return
				}

				proc := NewProcessor(d.Logger(), r.Context(), db)
				if err := proc.RemoveAlias(aliasID); err != nil {
					d.Logger().WithError(err).Error("Failed to remove alias")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}

func ingredientRecipesHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				page := queryInt(r, "page[number]", 1)
				pageSize := queryInt(r, "page[size]", 20)

				proc := NewProcessor(d.Logger(), r.Context(), db)
				refs, total, err := proc.GetIngredientRecipes(id, page, pageSize)
				if err != nil {
					d.Logger().WithError(err).Error("Failed to get ingredient recipes")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				w.Header().Set("Content-Type", "application/vnd.api+json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"data": refs,
					"meta": map[string]interface{}{
						"total":    total,
						"page":     page,
						"pageSize": pageSize,
					},
				})
			}
		})
	}
}

func reassignHandler(db *gorm.DB) server.InputHandler[ReassignRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input ReassignRequest) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				targetID, err := uuid.Parse(input.TargetIngredientId)
				if err != nil {
					server.WriteError(w, http.StatusBadRequest, "Invalid Request", "targetIngredientId must be a valid UUID")
					return
				}

				proc := NewProcessor(d.Logger(), r.Context(), db)
				reassigned, err := proc.Reassign(id, targetID)
				if err != nil {
					d.Logger().WithError(err).Error("Failed to reassign")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				if err := proc.Delete(id); err != nil {
					d.Logger().WithError(err).Error("Failed to delete after reassign")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				w.Header().Set("Content-Type", "application/vnd.api+json")
				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"meta": map[string]interface{}{
						"reassigned": reassigned,
					},
				})
			}
		})
	}
}

func bulkCategorizeHandler(db *gorm.DB) server.InputHandler[BulkCategorizeRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input BulkCategorizeRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			if len(input.IngredientIds) == 0 {
				server.WriteError(w, http.StatusUnprocessableEntity, "Validation Failed", "ingredient_ids must not be empty")
				return
			}
			if len(input.IngredientIds) > 200 {
				server.WriteError(w, http.StatusUnprocessableEntity, "Validation Failed", "ingredient_ids must not exceed 200 items")
				return
			}

			categoryID, err := uuid.Parse(input.CategoryId)
			if err != nil {
				server.WriteError(w, http.StatusUnprocessableEntity, "Validation Failed", "category_id must be a valid UUID")
				return
			}

			ingredientIDs := make([]uuid.UUID, len(input.IngredientIds))
			for i, idStr := range input.IngredientIds {
				parsed, err := uuid.Parse(idStr)
				if err != nil {
					server.WriteError(w, http.StatusUnprocessableEntity, "Validation Failed", "all ingredient_ids must be valid UUIDs")
					return
				}
				ingredientIDs[i] = parsed
			}

			proc := NewProcessor(d.Logger(), r.Context(), db)
			if err := proc.BulkCategorize(t.Id(), ingredientIDs, categoryID); err != nil {
				d.Logger().WithError(err).Error("Failed to bulk categorize ingredients")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			w.WriteHeader(http.StatusNoContent)
		}
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
