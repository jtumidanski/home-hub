package recipe

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/recipe-service/internal/recipe/cooklang"
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
		rihParse := server.RegisterInputHandler[ParseRequest](l)(si)
		rihRestore := server.RegisterInputHandler[RestorationRequest](l)(si)

		api.HandleFunc("/recipes/parse", rihParse("ParseRecipe", parseHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/recipes/tags", rh("ListRecipeTags", listTagsHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/recipes/restorations", rihRestore("RestoreRecipe", restoreHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/recipes", rh("ListRecipes", listHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/recipes", rihCreate("CreateRecipe", createHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/recipes/{id}", rh("GetRecipe", getHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/recipes/{id}", rihUpdate("UpdateRecipe", updateHandler(db))).Methods(http.MethodPatch)
		api.HandleFunc("/recipes/{id}", rh("DeleteRecipe", deleteHandler(db))).Methods(http.MethodDelete)
	}
}

func parseHandler(db *gorm.DB) server.InputHandler[ParseRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input ParseRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if len(input.Source) > cooklang.MaxSourceSize {
				server.WriteError(w, http.StatusBadRequest, "Source too large", fmt.Sprintf("Source must be under %d bytes", cooklang.MaxSourceSize))
				return
			}
			proc := NewProcessor(d.Logger(), r.Context(), db)
			result := proc.ParseSource(input.Source)

			rest := RestParseModel{
				Ingredients: result.Ingredients,
				Steps:       result.Steps,
				Metadata:    result.Metadata,
				Errors:      result.Errors,
			}
			if rest.Ingredients == nil {
				rest.Ingredients = []cooklang.Ingredient{}
			}
			if rest.Steps == nil {
				rest.Steps = []cooklang.Step{}
			}

			server.MarshalResponse[RestParseModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
		}
	}
}

func listHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			search := r.URL.Query().Get("search")
			tags := r.URL.Query()["tag"]
			page := queryInt(r, "page[number]", 1)
			pageSize := queryInt(r, "page[size]", 20)

			proc := NewProcessor(d.Logger(), r.Context(), db)
			models, total, err := proc.List(search, tags, page, pageSize)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to list recipes")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			rest := TransformListSlice(models)

			// Write custom response with pagination meta
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

			// Inject meta into response
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

func createHandler(db *gorm.DB) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())

			proc := NewProcessor(d.Logger(), r.Context(), db)
			m, parsed, err := proc.Create(t.Id(), t.HouseholdId(), CreateAttrs{
				Title:           input.Title,
				Description:     input.Description,
				Source:          input.Source,
				Servings:        input.Servings,
				PrepTimeMinutes: input.PrepTimeMinutes,
				CookTimeMinutes: input.CookTimeMinutes,
				SourceURL:       input.SourceURL,
				Tags:            input.Tags,
			})
			if err != nil {
				if errors.Is(err, ErrTitleRequired) || errors.Is(err, ErrSourceRequired) {
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
					return
				}
				if err.Error() == "invalid cooklang syntax" {
					writeCooklangErrors(w, parsed.Errors)
					return
				}
				d.Logger().WithError(err).Error("Failed to create recipe")
				server.WriteError(w, http.StatusInternalServerError, "Create Failed", "")
				return
			}

			rest := TransformDetail(m, parsed)
			server.MarshalCreatedResponse[RestDetailModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func getHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, parsed, err := proc.Get(id)
				if err != nil {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Recipe not found")
					return
				}
				rest := TransformDetail(m, parsed)
				server.MarshalResponse[RestDetailModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func updateHandler(db *gorm.DB) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input UpdateRequest) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				attrs := UpdateAttrs{}
				if input.Title != "" {
					attrs.Title = &input.Title
				}
				if input.Description != "" {
					attrs.Description = &input.Description
				}
				if input.Source != "" {
					attrs.Source = &input.Source
				}
				attrs.Servings = input.Servings
				attrs.PrepTimeMinutes = input.PrepTimeMinutes
				attrs.CookTimeMinutes = input.CookTimeMinutes
				if input.SourceURL != "" {
					attrs.SourceURL = &input.SourceURL
				}
				if input.Tags != nil {
					attrs.Tags = &input.Tags
				}

				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, parsed, err := proc.Update(id, attrs)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Recipe not found")
						return
					}
					if err.Error() == "invalid cooklang syntax" {
						writeCooklangErrors(w, parsed.Errors)
						return
					}
					d.Logger().WithError(err).Error("Failed to update recipe")
					server.WriteError(w, http.StatusInternalServerError, "Update Failed", "")
					return
				}

				rest := TransformDetail(m, parsed)
				server.MarshalResponse[RestDetailModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func deleteHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				proc := NewProcessor(d.Logger(), r.Context(), db)
				if err := proc.Delete(id); err != nil {
					d.Logger().WithError(err).Error("Failed to delete recipe")
					server.WriteError(w, http.StatusInternalServerError, "Delete Failed", "")
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}

func restoreHandler(db *gorm.DB) server.InputHandler[RestorationRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input RestorationRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			recipeID, err := uuid.Parse(input.RecipeId)
			if err != nil {
				server.WriteError(w, http.StatusBadRequest, "Invalid Request", "recipeId must be a valid UUID")
				return
			}

			proc := NewProcessor(d.Logger(), r.Context(), db)
			m, parsed, err := proc.Restore(recipeID)
			if err != nil {
				if errors.Is(err, ErrNotFound) {
					server.WriteError(w, http.StatusNotFound, "Not Found", "Recipe not found")
				} else if errors.Is(err, ErrNotDeleted) {
					server.WriteError(w, http.StatusBadRequest, "Bad Request", "Recipe is not deleted")
				} else if errors.Is(err, ErrRestoreWindow) {
					server.WriteError(w, http.StatusGone, "Gone", "Restore window expired")
				} else {
					d.Logger().WithError(err).Error("Failed to restore recipe")
					server.WriteError(w, http.StatusInternalServerError, "Restore Failed", "")
				}
				return
			}

			rest := TransformDetail(m, parsed)
			server.MarshalResponse[RestDetailModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
		}
	}
}

func listTagsHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			proc := NewProcessor(d.Logger(), r.Context(), db)
			tags, err := proc.ListTags()
			if err != nil {
				d.Logger().WithError(err).Error("Failed to list tags")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			rest := make([]RestTagModel, len(tags))
			for i, t := range tags {
				rest[i] = RestTagModel{Tag: t.Tag, Count: t.Count}
			}

			server.MarshalSliceResponse[RestTagModel](d.Logger())(w)(c.ServerInformation())(rest)
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

type cooklangErrorResponse struct {
	Status string            `json:"status"`
	Title  string            `json:"title"`
	Detail string            `json:"detail"`
	Source map[string]string `json:"source,omitempty"`
}

func writeCooklangErrors(w http.ResponseWriter, errs []cooklang.ParseError) {
	w.Header().Set("Content-Type", "application/vnd.api+json")
	w.WriteHeader(http.StatusUnprocessableEntity)

	apiErrors := make([]cooklangErrorResponse, len(errs))
	for i, e := range errs {
		apiErrors[i] = cooklangErrorResponse{
			Status: "422",
			Title:  "Invalid Cooklang syntax",
			Detail: e.Message,
			Source: map[string]string{"pointer": "/data/attributes/source"},
		}
	}
	json.NewEncoder(w).Encode(map[string]interface{}{"errors": apiErrors})
}
