package normalization

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/audit"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitializeRoutes(db *gorm.DB) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		rihResolve := server.RegisterInputHandler[ResolveRequest](l)(si)

		api.HandleFunc("/recipes/{id}/ingredients/{ingredientId}/resolve", rihResolve("ResolveIngredient", resolveHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/recipes/{id}/renormalize", rh("RenormalizeRecipe", renormalizeHandler(db))).Methods(http.MethodPost)
	}
}

func resolveHandler(db *gorm.DB) server.InputHandler[ResolveRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input ResolveRequest) http.HandlerFunc {
		return server.ParseID("id", func(recipeID uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				ingredientIDStr := mux.Vars(r)["ingredientId"]
				ingredientID, err := uuid.Parse(ingredientIDStr)
				if err != nil {
					server.WriteError(w, http.StatusBadRequest, "Invalid Request", "ingredientId must be a valid UUID")
					return
				}

				canonicalID, err := uuid.Parse(input.CanonicalIngredientId)
				if err != nil {
					server.WriteError(w, http.StatusBadRequest, "Invalid Request", "canonicalIngredientId must be a valid UUID")
					return
				}

				t := tenantctx.MustFromContext(r.Context())
				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.ResolveIngredient(ingredientID, canonicalID, input.SaveAsAlias)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Recipe ingredient not found")
						return
					}
					d.Logger().WithError(err).Error("Failed to resolve ingredient")
					server.WriteError(w, http.StatusInternalServerError, "Resolve Failed", "")
					return
				}

				audit.Emit(d.Logger(), db.WithContext(r.Context()), t.Id(), "recipe_ingredient", ingredientID, "normalization.corrected", t.UserId(), map[string]interface{}{
					"recipe_id":               recipeID,
					"canonical_ingredient_id": canonicalID,
					"save_as_alias":           input.SaveAsAlias,
				})

				if input.SaveAsAlias {
					audit.Emit(d.Logger(), db.WithContext(r.Context()), t.Id(), "canonical_ingredient", canonicalID, "ingredient.alias_created", t.UserId(), map[string]interface{}{
						"alias_name": m.RawName(),
					})
				}

				rest := TransformIngredient(m)
				server.MarshalResponse[RestIngredientModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func renormalizeHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(recipeID uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())
				proc := NewProcessor(d.Logger(), r.Context(), db)
				models, summary, err := proc.Renormalize(t.Id(), recipeID)
				if err != nil {
					d.Logger().WithError(err).Error("Failed to renormalize")
					server.WriteError(w, http.StatusInternalServerError, "Renormalize Failed", "")
					return
				}

				audit.Emit(d.Logger(), db.WithContext(r.Context()), t.Id(), "recipe", recipeID, "recipe.renormalized", t.UserId(), map[string]interface{}{
					"total":            summary.Total,
					"changed":          summary.Changed,
					"still_unresolved": summary.StillUnresolved,
				})

				rest := TransformIngredients(models)
				items := make([]jsonapi.MarshalIdentifier, len(rest))
				for i := range rest {
					items[i] = &rest[i]
				}
				result, _ := jsonapi.MarshalWithURLs(items, c.ServerInformation())

				w.Header().Set("Content-Type", "application/vnd.api+json")
				w.WriteHeader(http.StatusOK)
				w.Write(result)
			}
		})
	}
}
