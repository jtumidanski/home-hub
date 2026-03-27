package normalization

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
		return server.ParseID("id", func(_ uuid.UUID) http.HandlerFunc {
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

				proc := NewProcessor(d.Logger(), r.Context(), db)
				result, err := proc.ResolveIngredient(ingredientID, canonicalID, input.SaveAsAlias)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Recipe ingredient not found")
						return
					}
					d.Logger().WithError(err).Error("Failed to resolve ingredient")
					server.WriteError(w, http.StatusInternalServerError, "Resolve Failed", "")
					return
				}

				rest := TransformIngredient(result.Model)
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
				models, _, err := proc.Renormalize(t.Id(), recipeID)
				if err != nil {
					d.Logger().WithError(err).Error("Failed to renormalize")
					server.WriteError(w, http.StatusInternalServerError, "Renormalize Failed", "")
					return
				}

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
