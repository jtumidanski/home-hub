package user

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/packages/shared-go/auth"
	"github.com/jtumidanski/home-hub/packages/shared-go/model/ops"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/jtumidanski/home-hub/packages/shared-go/rest/server"
)

// InitializeRoutes registers all user-related routes
func InitializeRoutes(si jsonapi.ServerInformation) func(db *gorm.DB) server.RouteInitializer {
	return func(db *gorm.DB) server.RouteInitializer {
		// User CRUD endpoints
		// User-household relationship endpoints
		return func(router *mux.Router, l logrus.FieldLogger) {
			// Auth endpoint - /me returns the currently authenticated user
			router.HandleFunc("/me", server.RegisterHandler(l)(si)("get-me", getMeHandler(db))).Methods(http.MethodGet)

			// User CRUD
			router.HandleFunc("/users", server.RegisterHandler(l)(si)("get-users", listUsersHandler(db))).Methods(http.MethodGet)
			router.HandleFunc("/users", server.RegisterInputHandler[CreateRequest](l)(si)("create-user", createUserHandler(db))).Methods(http.MethodPost)
			router.HandleFunc("/users/{id}", server.RegisterHandler(l)(si)("get-user", getUserHandler(db))).Methods(http.MethodGet)
			router.HandleFunc("/users/{id}", server.RegisterInputHandler[UpdateRequest](l)(si)("update-user", updateUserHandler(db))).Methods(http.MethodPatch)
			router.HandleFunc("/users/{id}", server.RegisterHandler(l)(si)("delete-user", deleteUserHandler(db))).Methods(http.MethodDelete)
			router.HandleFunc("/users/{id}/relationships/household", server.RegisterInputHandler[AssociateHouseholdRequest](l)(si)("associate-household", associateHouseholdHandler(db))).Methods(http.MethodPost)
			router.HandleFunc("/users/{id}/relationships/household", server.RegisterHandler(l)(si)("disassociate-household", disassociateHouseholdHandler(db))).Methods(http.MethodDelete)
		}
	}
}

// IdHandler a handler interface which requires a userId
type IdHandler func(userId uuid.UUID) http.HandlerFunc

// ParseId parses the userId consistently from the request, and provide it to the next handler
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

// getMeHandler handles GET /me - returns the currently authenticated user with roles
func getMeHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Extract auth context
			authCtx, ok := auth.FromContext(r.Context())
			if !ok {
				d.Logger().Error("Auth context not found in /me handler")
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Get user by ID
			model, err := NewProcessor(d.Logger(), r.Context(), db).GetById(authCtx.UserId)()
			if err != nil {
				if errors.Is(err, ErrUserNotFound) {
					d.Logger().WithError(err).Error("Authenticated user not found in database")
					w.WriteHeader(http.StatusNotFound)
					return
				}
				d.Logger().WithError(err).Error("Failed to fetch authenticated user")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Transform to MeResponse with roles
			meResponse, err := TransformToMe(model, authCtx.Roles)
			if err != nil {
				d.Logger().WithError(err).Errorf("Creating MeResponse model.")
				w.WriteHeader(http.StatusInternalServerError)
				return
			}

			// Marshal using JSON:API format
			server.MarshalResponse[MeResponse](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(meResponse)
		}
	}
}

// listUsersHandler handles GET /users
func listUsersHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			models, err := GetAll(db)()
			if err != nil {
				d.Logger().WithError(err).Error("error listing users")
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

// getUserHandler handles GET /users/:id
func getUserHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return ParseId(d.Logger(), func(userId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				model, err := NewProcessor(d.Logger(), r.Context(), db).GetById(userId)()
				if err != nil {
					if errors.Is(err, ErrUserNotFound) {
						d.Logger().WithError(err).Error("User not found")
						w.WriteHeader(http.StatusNotFound)
						return
					}
					d.Logger().WithError(err).Error("Failed to fetch user")
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

// createUserHandler handles POST /users
func createUserHandler(db *gorm.DB) server.InputHandler[CreateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, req CreateRequest) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			input := CreateInput{
				Email:       req.Email,
				DisplayName: req.DisplayName,
				HouseholdId: req.HouseholdId,
			}

			model, err := NewProcessor(d.Logger(), r.Context(), db).Create(input)()
			if err != nil {
				if errors.Is(err, ErrEmailAlreadyExists) {
					d.Logger().WithError(err).Errorf("Email already exists.")
					w.WriteHeader(http.StatusConflict)
					return
				}
				if errors.Is(err, ErrHouseholdNotFound) {
					d.Logger().WithError(err).Errorf("Household not found.")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				if errors.Is(err, ErrEmailRequired) || errors.Is(err, ErrEmailInvalid) ||
					errors.Is(err, ErrDisplayNameRequired) || errors.Is(err, ErrDisplayNameEmpty) {
					d.Logger().WithError(err).Errorf("Validation failed.")
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				d.Logger().WithError(err).Errorf("Failed to create user.")
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

// updateUserHandler handles PATCH /users/:id
func updateUserHandler(db *gorm.DB) server.InputHandler[UpdateRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, req UpdateRequest) http.HandlerFunc {
		return ParseId(d.Logger(), func(userId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				input := UpdateInput{
					Email:       req.Email,
					DisplayName: req.DisplayName,
				}

				model, err := NewProcessor(d.Logger(), r.Context(), db).Update(userId, input)()
				if err != nil {
					if errors.Is(err, ErrUserNotFound) {
						d.Logger().WithError(err).Errorf("User not found.")
						w.WriteHeader(http.StatusNotFound)
						return
					}
					if errors.Is(err, ErrEmailAlreadyExists) {
						d.Logger().WithError(err).Errorf("Email already exists.")
						w.WriteHeader(http.StatusConflict)
						return
					}
					if errors.Is(err, ErrEmailRequired) || errors.Is(err, ErrEmailInvalid) ||
						errors.Is(err, ErrDisplayNameRequired) || errors.Is(err, ErrDisplayNameEmpty) {
						d.Logger().WithError(err).Errorf("Validation failed.")
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					d.Logger().WithError(err).Errorf("Failed to update user.")
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

// deleteUserHandler handles DELETE /users/:id
func deleteUserHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return ParseId(d.Logger(), func(userId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				err := NewProcessor(d.Logger(), r.Context(), db).Delete(userId)
				if err != nil {
					if errors.Is(err, ErrUserNotFound) {
						d.Logger().WithError(err).Error("User not found")
						w.WriteHeader(http.StatusNotFound)
						return
					}
					d.Logger().WithError(err).Error("Failed to delete user")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}

// associateHouseholdHandler handles POST /users/:id/relationships/household
func associateHouseholdHandler(db *gorm.DB) server.InputHandler[AssociateHouseholdRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, req AssociateHouseholdRequest) http.HandlerFunc {
		return ParseId(d.Logger(), func(userId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				model, err := NewProcessor(d.Logger(), d.Context(), db).AssociateHousehold(userId, req.Id)()
				if err != nil {
					if errors.Is(err, ErrUserNotFound) {
						d.Logger().WithError(err).Error("User not found")
						w.WriteHeader(http.StatusNotFound)
						return
					}
					if errors.Is(err, ErrHouseholdNotFound) {
						d.Logger().WithError(err).Error("Household not found")
						w.WriteHeader(http.StatusNotFound)
						return
					}
					d.Logger().WithError(err).Error("Failed to associate household")
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

// disassociateHouseholdHandler handles DELETE /users/:id/relationships/household
func disassociateHouseholdHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return ParseId(d.Logger(), func(userId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				model, err := NewProcessor(d.Logger(), r.Context(), db).DisassociateHousehold(userId)()
				if err != nil {
					if errors.Is(err, ErrUserNotFound) {
						d.Logger().WithError(err).Error("User not found")
						w.WriteHeader(http.StatusNotFound)
						return
					}
					d.Logger().WithError(err).Error("Failed to disassociate household")
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
