package list

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/services/shopping-service/internal/categoryclient"
	"github.com/jtumidanski/home-hub/services/shopping-service/internal/item"
	"github.com/jtumidanski/home-hub/services/shopping-service/internal/recipeclient"
	"github.com/jtumidanski/home-hub/shared/go/server"
	tenantctx "github.com/jtumidanski/home-hub/shared/go/tenant"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

func InitializeRoutes(db *gorm.DB, categoryServiceURL, recipeServiceURL string) func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
	catClient := categoryclient.New(categoryServiceURL)
	recipeClient := recipeclient.New(recipeServiceURL)

	return func(l logrus.FieldLogger, si jsonapi.ServerInformation, api *mux.Router) {
		rh := server.RegisterHandler(l)(si)
		rihCreate := server.RegisterInputHandler[CreateRequest](l)(si)
		rihUpdate := server.RegisterInputHandler[UpdateRequest](l)(si)
		rihArchive := server.RegisterInputHandler[ArchiveRequest](l)(si)
		rihUnarchive := server.RegisterInputHandler[UnarchiveRequest](l)(si)
		rihItemCreate := server.RegisterInputHandler[item.CreateRequest](l)(si)
		rihItemUpdate := server.RegisterInputHandler[item.UpdateRequest](l)(si)
		rihItemCheck := server.RegisterInputHandler[item.CheckRequest](l)(si)
		rihUncheckAll := server.RegisterInputHandler[UncheckAllRequest](l)(si)
		rihImport := server.RegisterInputHandler[ImportRequest](l)(si)

		api.HandleFunc("/shopping/lists", rh("ListShoppingLists", listHandler(db, catClient, recipeClient))).Methods(http.MethodGet)
		api.HandleFunc("/shopping/lists", rihCreate("CreateShoppingList", createHandler(db, catClient, recipeClient))).Methods(http.MethodPost)
		api.HandleFunc("/shopping/lists/{id}", rh("GetShoppingList", getHandler(db, catClient, recipeClient))).Methods(http.MethodGet)
		api.HandleFunc("/shopping/lists/{id}", rihUpdate("UpdateShoppingList", updateHandler(db, catClient, recipeClient))).Methods(http.MethodPatch)
		api.HandleFunc("/shopping/lists/{id}", rh("DeleteShoppingList", deleteHandler(db, catClient, recipeClient))).Methods(http.MethodDelete)
		api.HandleFunc("/shopping/lists/{id}/archive", rihArchive("ArchiveShoppingList", archiveHandler(db, catClient, recipeClient))).Methods(http.MethodPost)
		api.HandleFunc("/shopping/lists/{id}/unarchive", rihUnarchive("UnarchiveShoppingList", unarchiveHandler(db, catClient, recipeClient))).Methods(http.MethodPost)

		api.HandleFunc("/shopping/lists/{id}/items", rihItemCreate("AddShoppingItem", addItemHandler(db, catClient, recipeClient))).Methods(http.MethodPost)
		api.HandleFunc("/shopping/lists/{id}/items/{itemId}", rihItemUpdate("UpdateShoppingItem", updateItemHandler(db, catClient, recipeClient))).Methods(http.MethodPatch)
		api.HandleFunc("/shopping/lists/{id}/items/{itemId}", rh("RemoveShoppingItem", removeItemHandler(db, catClient, recipeClient))).Methods(http.MethodDelete)
		api.HandleFunc("/shopping/lists/{id}/items/{itemId}/check", rihItemCheck("CheckShoppingItem", checkItemHandler(db, catClient, recipeClient))).Methods(http.MethodPatch)
		api.HandleFunc("/shopping/lists/{id}/items/uncheck-all", rihUncheckAll("UncheckAllItems", uncheckAllHandler(db, catClient, recipeClient))).Methods(http.MethodPost)

		api.HandleFunc("/shopping/lists/{id}/import/meal-plan", rihImport("ImportMealPlan", importHandler(db, catClient, recipeClient))).Methods(http.MethodPost)
	}
}

// accessTokenCookie extracts the access_token cookie value for forwarding to
// downstream services that share the same auth middleware. Returns "" when the
// cookie is missing.
func accessTokenCookie(r *http.Request) string {
	if c, err := r.Cookie("access_token"); err == nil {
		return c.Value
	}
	return ""
}

func listHandler(db *gorm.DB, catClient *categoryclient.Client, recipeClient *recipeclient.Client) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			status := r.URL.Query().Get("status")
			proc := NewProcessor(d.Logger(), r.Context(), db, catClient, recipeClient)

			models, err := proc.List(status)
			if err != nil {
				d.Logger().WithError(err).Error("Failed to list shopping lists")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			transformed, err := TransformSlice(models)
			if err != nil {
				d.Logger().WithError(err).Error("Creating REST models.")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			rest := make([]*RestModel, len(transformed))
			for i := range transformed {
				rest[i] = &transformed[i]
			}
			server.MarshalSliceResponse[*RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func createHandler(db *gorm.DB, catClient *categoryclient.Client, recipeClient *recipeclient.Client) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db, catClient, recipeClient)

			m, err := proc.Create(t.Id(), t.HouseholdId(), t.UserId(), input.Name)
			if err != nil {
				if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrNameTooLong) {
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
					return
				}
				d.Logger().WithError(err).Error("Failed to create shopping list")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}

			rest, err := Transform(m)
			if err != nil {
				d.Logger().WithError(err).Error("Creating REST model.")
				server.WriteError(w, http.StatusInternalServerError, "Error", "")
				return
			}
			server.MarshalCreatedResponse[RestModel](d.Logger())(w)(c.ServerInformation())(rest)
		}
	}
}

func getHandler(db *gorm.DB, catClient *categoryclient.Client, recipeClient *recipeclient.Client) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				proc := NewProcessor(d.Logger(), r.Context(), db, catClient, recipeClient)
				m, items, err := proc.GetWithItems(id)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Shopping list not found")
						return
					}
					d.Logger().WithError(err).Error("Failed to get shopping list")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				restItems := item.TransformNestedSlice(items)
				rest, err := TransformWithItems(m, restItems)
				if err != nil {
					d.Logger().WithError(err).Error("Creating REST model.")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func updateHandler(db *gorm.DB, catClient *categoryclient.Client, recipeClient *recipeclient.Client) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input UpdateRequest) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				proc := NewProcessor(d.Logger(), r.Context(), db, catClient, recipeClient)

				m, err := proc.Update(id, input.Name)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Shopping list not found")
						return
					}
					if errors.Is(err, ErrArchived) {
						server.WriteError(w, http.StatusConflict, "Conflict", "Cannot modify archived list")
						return
					}
					if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrNameTooLong) {
						server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
						return
					}
					d.Logger().WithError(err).Error("Failed to update shopping list")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				rest, err := Transform(m)
				if err != nil {
					d.Logger().WithError(err).Error("Creating REST model.")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func deleteHandler(db *gorm.DB, catClient *categoryclient.Client, recipeClient *recipeclient.Client) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				proc := NewProcessor(d.Logger(), r.Context(), db, catClient, recipeClient)
				if err := proc.Delete(id); err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Shopping list not found")
						return
					}
					d.Logger().WithError(err).Error("Failed to delete shopping list")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}

func archiveHandler(db *gorm.DB, catClient *categoryclient.Client, recipeClient *recipeclient.Client) server.InputHandler[ArchiveRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, _ ArchiveRequest) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				proc := NewProcessor(d.Logger(), r.Context(), db, catClient, recipeClient)
				m, err := proc.Archive(id)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Shopping list not found")
						return
					}
					if errors.Is(err, ErrAlreadyArchived) {
						server.WriteError(w, http.StatusConflict, "Conflict", "List is already archived")
						return
					}
					d.Logger().WithError(err).Error("Failed to archive shopping list")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				rest, err := Transform(m)
				if err != nil {
					d.Logger().WithError(err).Error("Creating REST model.")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func unarchiveHandler(db *gorm.DB, catClient *categoryclient.Client, recipeClient *recipeclient.Client) server.InputHandler[UnarchiveRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, _ UnarchiveRequest) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				proc := NewProcessor(d.Logger(), r.Context(), db, catClient, recipeClient)
				m, err := proc.Unarchive(id)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Shopping list not found")
						return
					}
					if errors.Is(err, ErrNotArchived) {
						server.WriteError(w, http.StatusConflict, "Conflict", "List is not archived")
						return
					}
					d.Logger().WithError(err).Error("Failed to unarchive shopping list")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				rest, err := Transform(m)
				if err != nil {
					d.Logger().WithError(err).Error("Creating REST model.")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func addItemHandler(db *gorm.DB, catClient *categoryclient.Client, recipeClient *recipeclient.Client) server.InputHandler[item.CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input item.CreateRequest) http.HandlerFunc {
		return server.ParseID("id", func(listID uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				addInput := item.AddInput{
					ListID:     listID,
					Name:       input.Name,
					Quantity:   input.Quantity,
					CategoryID: input.CategoryId,
					Position:   input.Position,
				}

				proc := NewProcessor(d.Logger(), r.Context(), db, catClient, recipeClient)
				m, err := proc.AddItem(listID, addInput, accessTokenCookie(r))
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Shopping list not found")
						return
					}
					if errors.Is(err, ErrArchived) {
						server.WriteError(w, http.StatusConflict, "Conflict", "Cannot modify archived list")
						return
					}
					if errors.Is(err, item.ErrNameRequired) || errors.Is(err, item.ErrNameTooLong) {
						server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
						return
					}
					d.Logger().WithError(err).Error("Failed to add shopping item")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				rest, err := item.Transform(m)
				if err != nil {
					d.Logger().WithError(err).Error("Creating REST model.")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				server.MarshalCreatedResponse[item.RestModel](d.Logger())(w)(c.ServerInformation())(rest)
			}
		})
	}
}

func updateItemHandler(db *gorm.DB, catClient *categoryclient.Client, recipeClient *recipeclient.Client) server.InputHandler[item.UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input item.UpdateRequest) http.HandlerFunc {
		return server.ParseID("id", func(listID uuid.UUID) http.HandlerFunc {
			return server.ParseID("itemId", func(itemID uuid.UUID) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					updateInput := item.UpdateInput{
						Name:       input.Name,
						Quantity:   input.Quantity,
						CategoryID: input.CategoryId,
						Position:   input.Position,
					}

					proc := NewProcessor(d.Logger(), r.Context(), db, catClient, recipeClient)
					m, err := proc.UpdateItem(listID, itemID, updateInput, accessTokenCookie(r))
					if err != nil {
						if errors.Is(err, ErrNotFound) {
							server.WriteError(w, http.StatusNotFound, "Not Found", "Shopping list not found")
							return
						}
						if errors.Is(err, ErrArchived) {
							server.WriteError(w, http.StatusConflict, "Conflict", "Cannot modify archived list")
							return
						}
						if errors.Is(err, item.ErrNotFound) {
							server.WriteError(w, http.StatusNotFound, "Not Found", "Item not found")
							return
						}
						if errors.Is(err, item.ErrNameRequired) || errors.Is(err, item.ErrNameTooLong) {
							server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
							return
						}
						d.Logger().WithError(err).Error("Failed to update shopping item")
						server.WriteError(w, http.StatusInternalServerError, "Error", "")
						return
					}

					rest, err := item.Transform(m)
					if err != nil {
						d.Logger().WithError(err).Error("Creating REST model.")
						server.WriteError(w, http.StatusInternalServerError, "Error", "")
						return
					}
					server.MarshalResponse[item.RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
				}
			})
		})
	}
}

func removeItemHandler(db *gorm.DB, catClient *categoryclient.Client, recipeClient *recipeclient.Client) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(listID uuid.UUID) http.HandlerFunc {
			return server.ParseID("itemId", func(itemID uuid.UUID) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					proc := NewProcessor(d.Logger(), r.Context(), db, catClient, recipeClient)
					if err := proc.RemoveItem(listID, itemID); err != nil {
						if errors.Is(err, ErrNotFound) {
							server.WriteError(w, http.StatusNotFound, "Not Found", "Shopping list not found")
							return
						}
						if errors.Is(err, ErrArchived) {
							server.WriteError(w, http.StatusConflict, "Conflict", "Cannot modify archived list")
							return
						}
						if errors.Is(err, item.ErrNotFound) {
							server.WriteError(w, http.StatusNotFound, "Not Found", "Item not found")
							return
						}
						d.Logger().WithError(err).Error("Failed to remove shopping item")
						server.WriteError(w, http.StatusInternalServerError, "Error", "")
						return
					}
					w.WriteHeader(http.StatusNoContent)
				}
			})
		})
	}
}

func checkItemHandler(db *gorm.DB, catClient *categoryclient.Client, recipeClient *recipeclient.Client) server.InputHandler[item.CheckRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input item.CheckRequest) http.HandlerFunc {
		return server.ParseID("id", func(listID uuid.UUID) http.HandlerFunc {
			return server.ParseID("itemId", func(itemID uuid.UUID) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					proc := NewProcessor(d.Logger(), r.Context(), db, catClient, recipeClient)
					m, err := proc.CheckItem(listID, itemID, input.Checked)
					if err != nil {
						if errors.Is(err, ErrNotFound) {
							server.WriteError(w, http.StatusNotFound, "Not Found", "Shopping list not found")
							return
						}
						if errors.Is(err, ErrArchived) {
							server.WriteError(w, http.StatusConflict, "Conflict", "Cannot modify archived list")
							return
						}
						if errors.Is(err, item.ErrNotFound) {
							server.WriteError(w, http.StatusNotFound, "Not Found", "Item not found")
							return
						}
						d.Logger().WithError(err).Error("Failed to check shopping item")
						server.WriteError(w, http.StatusInternalServerError, "Error", "")
						return
					}

					rest, err := item.Transform(m)
					if err != nil {
						d.Logger().WithError(err).Error("Creating REST model.")
						server.WriteError(w, http.StatusInternalServerError, "Error", "")
						return
					}
					server.MarshalResponse[item.RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
				}
			})
		})
	}
}

func uncheckAllHandler(db *gorm.DB, catClient *categoryclient.Client, recipeClient *recipeclient.Client) server.InputHandler[UncheckAllRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, _ UncheckAllRequest) http.HandlerFunc {
		return server.ParseID("id", func(listID uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				proc := NewProcessor(d.Logger(), r.Context(), db, catClient, recipeClient)
				m, items, err := proc.UncheckAllItems(listID)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Shopping list not found")
						return
					}
					if errors.Is(err, ErrArchived) {
						server.WriteError(w, http.StatusConflict, "Conflict", "Cannot modify archived list")
						return
					}
					d.Logger().WithError(err).Error("Failed to uncheck all items")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				restItems := item.TransformNestedSlice(items)
				rest, err := TransformWithItems(m, restItems)
				if err != nil {
					d.Logger().WithError(err).Error("Creating REST model.")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}

func importHandler(db *gorm.DB, catClient *categoryclient.Client, recipeClient *recipeclient.Client) server.InputHandler[ImportRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input ImportRequest) http.HandlerFunc {
		return server.ParseID("id", func(listID uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				if input.PlanId == uuid.Nil {
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", "plan_id is required")
					return
				}

				proc := NewProcessor(d.Logger(), r.Context(), db, catClient, recipeClient)
				m, items, importedCount, err := proc.ImportFromMealPlan(listID, input.PlanId, accessTokenCookie(r))
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Shopping list not found")
						return
					}
					if errors.Is(err, ErrArchived) {
						server.WriteError(w, http.StatusConflict, "Conflict", "Cannot modify archived list")
						return
					}
					d.Logger().WithError(err).Error("Failed to import meal plan")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}

				restItems := item.TransformNestedSlice(items)
				rest, err := TransformImported(m, restItems, importedCount)
				if err != nil {
					d.Logger().WithError(err).Error("Creating REST model.")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(rest)
			}
		})
	}
}
