package household

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/packages/shared-go/model/ops"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	userPkg "github.com/jtumidanski/home-hub/apps/svc-users/user"
	"github.com/jtumidanski/home-hub/packages/shared-go/rest/server"
)

// InitializeRoutes registers all household-related routes
func InitializeRoutes(si jsonapi.ServerInformation) func(db *gorm.DB) server.RouteInitializer {
	return func(db *gorm.DB) server.RouteInitializer {
		return func(router *mux.Router, l logrus.FieldLogger) {
			// Household CRUD endpoints
			router.HandleFunc("/households", server.RegisterHandler(l)(si)("get-households", listHouseholdsHandler(db))).Methods(http.MethodGet)
			router.HandleFunc("/households", server.RegisterInputHandler[CreateRequest](l)(si)("create-household", createHouseholdHandler(db))).Methods(http.MethodPost)
			router.HandleFunc("/households/count", server.RegisterHandler(l)(si)("count-households", countHouseholdsHandler(db))).Methods(http.MethodGet)
			router.HandleFunc("/households/{id}", server.RegisterHandler(l)(si)("get-household", getHouseholdHandler(db))).Methods(http.MethodGet)
			router.HandleFunc("/households/{id}", server.RegisterInputHandler[UpdateRequest](l)(si)("update-household", updateHouseholdHandler(db))).Methods(http.MethodPatch)
			router.HandleFunc("/households/{id}", server.RegisterHandler(l)(si)("delete-household", deleteHouseholdHandler(db))).Methods(http.MethodDelete)

			// Household relationships
			router.HandleFunc("/households/{id}/users", server.RegisterHandler(l)(si)("get-household-users", getHouseholdUsersHandler(db))).Methods(http.MethodGet)
		}
	}
}

// IdHandler a handler interface which requires a householdId
type IdHandler func(householdId uuid.UUID) http.HandlerFunc

// ParseId parses the householdId consistently from the request, and provide it to the next handler
func ParseId(l logrus.FieldLogger, next IdHandler) http.HandlerFunc {
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

// listHouseholdsHandler handles GET /households
func listHouseholdsHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			models, err := GetAll(db)()
			if err != nil {
				d.Logger().WithError(err).Error("error listing households")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			res, err := ops.SliceMap(Transform)(ops.FixedProvider(models))(ops.ParallelMap())()
			if err != nil {
				d.Logger().WithError(err).Errorf("Creating REST model.")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			server.MarshalResponse[[]RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(res)
		}
	}
}

// getHouseholdHandler handles GET /households/:id
func getHouseholdHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return ParseId(d.Logger(), func(householdId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				model, err := NewProcessor(d.Logger(), r.Context(), db).GetById(householdId)()
				if err != nil {
					if errors.Is(err, ErrHouseholdNotFound) {
						d.Logger().WithError(err).Error("Household not found")
						w.WriteHeader(http.StatusNotFound)
						return
					}
					d.Logger().WithError(err).Error("Failed to fetch household")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				res, err := ops.Map(Transform)(ops.FixedProvider(model))()
				if err != nil {
					d.Logger().WithError(err).Errorf("Creating REST model.")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(res)
			}
		})
	}
}

// createHouseholdHandler handles POST /households
func createHouseholdHandler(db *gorm.DB) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, req CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			input := CreateInput{
				Name: req.Name,
			}

			model, err := NewProcessor(d.Logger(), r.Context(), db).Create(input)()
			if err != nil {
				if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrNameEmpty) {
					d.Logger().WithError(err).Errorf("Validation failed.")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				d.Logger().WithError(err).Errorf("Failed to create household.")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			res, err := ops.Map(Transform)(ops.FixedProvider(model))()
			if err != nil {
				d.Logger().WithError(err).Errorf("Creating REST model.")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(res)
		}
	}
}

// updateHouseholdHandler handles PATCH /households/:id
func updateHouseholdHandler(db *gorm.DB) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, req UpdateRequest) http.HandlerFunc {
		return ParseId(d.Logger(), func(householdId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				input := UpdateInput{
					Name: req.Name,
				}

				model, err := NewProcessor(d.Logger(), r.Context(), db).Update(householdId, input)()
				if err != nil {
					if errors.Is(err, ErrHouseholdNotFound) {
						d.Logger().WithError(err).Errorf("Household not found.")
						w.WriteHeader(http.StatusNotFound)
						return
					}
					if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrNameEmpty) {
						d.Logger().WithError(err).Errorf("Validation failed.")
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					d.Logger().WithError(err).Errorf("Failed to update household.")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				res, err := ops.Map(Transform)(ops.FixedProvider(model))()
				if err != nil {
					d.Logger().WithError(err).Errorf("Creating REST model.")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(res)
			}
		})
	}
}

// deleteHouseholdHandler handles DELETE /households/:id
func deleteHouseholdHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return ParseId(d.Logger(), func(householdId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				err := NewProcessor(d.Logger(), r.Context(), db).Delete(householdId)
				if err != nil {
					if errors.Is(err, ErrHouseholdNotFound) {
						d.Logger().WithError(err).Error("Household not found")
						w.WriteHeader(http.StatusNotFound)
						return
					}
					if errors.Is(err, ErrHouseholdHasUsers) {
						d.Logger().WithError(err).Error("Cannot delete household with users")
						w.WriteHeader(http.StatusConflict)
						return
					}
					d.Logger().WithError(err).Error("Failed to delete household")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}

// countHouseholdsHandler handles GET /households/count
func countHouseholdsHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			count, err := Count(db)()
			if err != nil {
				d.Logger().WithError(err).Error("Failed to count households")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Return JSON:API response with meta
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(`{"meta":{"count":%d}}`, count)))
		}
	}
}

// getHouseholdUsersHandler handles GET /households/:id/users
func getHouseholdUsersHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return ParseId(d.Logger(), func(householdId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				// Verify household exists
				_, err := NewProcessor(d.Logger(), r.Context(), db).GetById(householdId)()
				if err != nil {
					if errors.Is(err, ErrHouseholdNotFound) {
						d.Logger().WithError(err).Error("Household not found")
						w.WriteHeader(http.StatusNotFound)
						return
					}
					d.Logger().WithError(err).Error("Failed to fetch household")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				// Get users in household
				users, err := userPkg.GetByHouseholdId(db)(householdId)()
				if err != nil {
					d.Logger().WithError(err).Error("Failed to fetch users")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				res, err := ops.SliceMap(userPkg.Transform)(ops.FixedProvider(users))(ops.ParallelMap())()
				if err != nil {
					d.Logger().WithError(err).Errorf("Creating REST model.")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				server.MarshalResponse[[]userPkg.RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(res)
			}
		})
	}
}
