package wishlist

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
		rihCreate := server.RegisterInputHandler[CreateRequest](l)(si)
		rihUpdate := server.RegisterInputHandler[UpdateRequest](l)(si)
		rihVote := server.RegisterInputHandler[VoteRequest](l)(si)

		api.HandleFunc("/shopping/wish-list/items", rh("ListWishListItems", listHandler(db))).Methods(http.MethodGet)
		api.HandleFunc("/shopping/wish-list/items", rihCreate("CreateWishListItem", createHandler(db))).Methods(http.MethodPost)
		api.HandleFunc("/shopping/wish-list/items/{id}", rihUpdate("UpdateWishListItem", updateHandler(db))).Methods(http.MethodPatch)
		api.HandleFunc("/shopping/wish-list/items/{id}", rh("DeleteWishListItem", deleteHandler(db))).Methods(http.MethodDelete)
		api.HandleFunc("/shopping/wish-list/items/{id}/vote", rihVote("VoteWishListItem", voteHandler(db))).Methods(http.MethodPost)
	}
}

func listHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db)
			models, err := proc.List(t.HouseholdId())
			if err != nil {
				d.Logger().WithError(err).Error("Failed to list wish list items")
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

func createHandler(db *gorm.DB) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			t := tenantctx.MustFromContext(r.Context())
			proc := NewProcessor(d.Logger(), r.Context(), db)

			m, err := proc.Create(t.Id(), t.HouseholdId(), t.UserId(), CreateInput{
				Name:             input.Name,
				PurchaseLocation: input.PurchaseLocation,
				Urgency:          input.Urgency,
			})
			if err != nil {
				if isValidationError(err) {
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
					return
				}
				d.Logger().WithError(err).Error("Failed to create wish list item")
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

func updateHandler(db *gorm.DB) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, input UpdateRequest) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				if input.VoteCount != nil {
					server.WriteError(w, http.StatusBadRequest, "Validation Failed", "vote_count cannot be modified via this endpoint; use POST /vote")
					return
				}
				t := tenantctx.MustFromContext(r.Context())
				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.Update(t.HouseholdId(), id, UpdateInput{
					Name:             input.Name,
					PurchaseLocation: input.PurchaseLocation,
					Urgency:          input.Urgency,
				})
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Wish list item not found")
						return
					}
					if isValidationError(err) {
						server.WriteError(w, http.StatusBadRequest, "Validation Failed", err.Error())
						return
					}
					d.Logger().WithError(err).Error("Failed to update wish list item")
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

func deleteHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())
				proc := NewProcessor(d.Logger(), r.Context(), db)
				if err := proc.Delete(t.HouseholdId(), id); err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Wish list item not found")
						return
					}
					d.Logger().WithError(err).Error("Failed to delete wish list item")
					server.WriteError(w, http.StatusInternalServerError, "Error", "")
					return
				}
				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}

func voteHandler(db *gorm.DB) server.InputHandler[VoteRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, _ VoteRequest) http.HandlerFunc {
		return server.ParseID("id", func(id uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				t := tenantctx.MustFromContext(r.Context())
				proc := NewProcessor(d.Logger(), r.Context(), db)
				m, err := proc.Vote(t.HouseholdId(), id)
				if err != nil {
					if errors.Is(err, ErrNotFound) {
						server.WriteError(w, http.StatusNotFound, "Not Found", "Wish list item not found")
						return
					}
					d.Logger().WithError(err).Error("Failed to vote on wish list item")
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

func isValidationError(err error) bool {
	return errors.Is(err, ErrNameRequired) ||
		errors.Is(err, ErrNameTooLong) ||
		errors.Is(err, ErrPurchaseLocationTooLong) ||
		errors.Is(err, ErrInvalidUrgency) ||
		errors.Is(err, ErrVoteCountNegative)
}
