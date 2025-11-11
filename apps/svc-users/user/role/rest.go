package role

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/packages/shared-go/rest/server"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// InitializeRoutes registers all role-related routes
func InitializeRoutes(si jsonapi.ServerInformation) func(db *gorm.DB) server.RouteInitializer {
	return func(db *gorm.DB) server.RouteInitializer {
		return func(router *mux.Router, l logrus.FieldLogger) {
			// Role endpoints - nested under /users/:userId/roles
			router.HandleFunc("/users/{userId}/roles", server.RegisterHandler(l)(si)("list-user-roles", listUserRolesHandler(db))).Methods(http.MethodGet)
			router.HandleFunc("/users/{userId}/roles", server.RegisterInputHandler[AddRoleRequest](l)(si)("add-user-role", addUserRoleHandler(db))).Methods(http.MethodPost)
			router.HandleFunc("/users/{userId}/roles/{role}", server.RegisterHandler(l)(si)("remove-user-role", removeUserRoleHandler(db))).Methods(http.MethodDelete)
		}
	}
}

// UserIdHandler a handler interface which requires a userId
type UserIdHandler func(userId uuid.UUID) http.HandlerFunc

// ParseUserId parses the userId consistently from the request, and provide it to the next handler
func ParseUserId(l logrus.FieldLogger, next UserIdHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userId, err := uuid.Parse(vars["userId"])
		if err != nil {
			l.WithError(err).Errorf("Unable to properly parse userId from path.")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		next(userId)(w, r)
	}
}

// listUserRolesHandler handles GET /users/:userId/roles
func listUserRolesHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return ParseUserId(d.Logger(), func(userId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				// Get all roles for the user
				models, err := GetByUserId(userId)(db)
				if err != nil {
					d.Logger().WithError(err).Error("Failed to fetch user roles")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				// Transform to REST models
				restModels, err := TransformSlice(models)
				if err != nil {
					d.Logger().WithError(err).Error("Failed to transform role models")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				server.MarshalResponse[[]RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(restModels)
			}
		})
	}
}

// addUserRoleHandler handles POST /users/:userId/roles
func addUserRoleHandler(db *gorm.DB) server.InputHandler[AddRoleRequest] {
	return func(d *server.HandlerDependency, c *server.HandlerContext, req AddRoleRequest) http.HandlerFunc {
		return ParseUserId(d.Logger(), func(userId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				// Validate role name
				if !ValidRoles[req.Role] {
					d.Logger().Errorf("Invalid role name: %s", req.Role)
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				// Assign the role
				model, err := AssignRole(userId, req.Role)(db)
				if err != nil {
					if errors.Is(err, ErrRoleAlreadyAssigned) {
						d.Logger().WithError(err).Errorf("Role already assigned")
						w.WriteHeader(http.StatusConflict)
						return
					}
					d.Logger().WithError(err).Error("Failed to assign role")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				// Transform to REST model
				restModel, err := Transform(model)
				if err != nil {
					d.Logger().WithError(err).Error("Failed to transform role model")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				server.MarshalResponse[RestModel](d.Logger())(w)(c.ServerInformation())(map[string][]string{})(restModel)
			}
		})
	}
}

// removeUserRoleHandler handles DELETE /users/:userId/roles/:role
func removeUserRoleHandler(db *gorm.DB) server.GetHandler {
	return func(d *server.HandlerDependency, c *server.HandlerContext) http.HandlerFunc {
		return ParseUserId(d.Logger(), func(userId uuid.UUID) http.HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) {
				vars := mux.Vars(r)
				roleName := vars["role"]

				// Validate role name
				if !ValidRoles[roleName] {
					d.Logger().Errorf("Invalid role name: %s", roleName)
					w.WriteHeader(http.StatusBadRequest)
					return
				}

				// Remove the role
				err := RemoveRole(userId, roleName)(db)
				if err != nil {
					if errors.Is(err, ErrRoleNotAssigned) {
						d.Logger().WithError(err).Error("Role not assigned")
						w.WriteHeader(http.StatusNotFound)
						return
					}
					d.Logger().WithError(err).Error("Failed to remove role")
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.WriteHeader(http.StatusNoContent)
			}
		})
	}
}
