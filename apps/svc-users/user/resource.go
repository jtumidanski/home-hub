package user

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	"github.com/jtumidanski/home-hub/packages/shared-go/rest/server"
)

// InitializeRoutes registers all user-related routes
func InitializeRoutes(db *gorm.DB) server.RouteInitializer {
	return func(router *mux.Router, l logrus.FieldLogger) {
		// User CRUD endpoints
		router.HandleFunc("/users", listUsersHandler(db, l)).Methods(http.MethodGet)
		router.HandleFunc("/users", createUserHandler(db, l)).Methods(http.MethodPost)
		router.HandleFunc("/users/{id}", getUserHandler(db, l)).Methods(http.MethodGet)
		router.HandleFunc("/users/{id}", updateUserHandler(db, l)).Methods(http.MethodPatch)
		router.HandleFunc("/users/{id}", deleteUserHandler(db, l)).Methods(http.MethodDelete)

		// User-household relationship endpoints
		router.HandleFunc("/users/{id}/relationships/household", associateHouseholdHandler(db, l)).Methods(http.MethodPost)
		router.HandleFunc("/users/{id}/relationships/household", disassociateHouseholdHandler(db, l)).Methods(http.MethodDelete)
	}
}

// listUsersHandler handles GET /users
func listUsersHandler(db *gorm.DB, l logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		models, err := GetAll(db)()
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to fetch users", err.Error())
			return
		}

		restModels, err := TransformSlice(models)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to transform users", err.Error())
			return
		}

		writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": restModels})
	}
}

// getUserHandler handles GET /users/:id
func getUserHandler(db *gorm.DB, l logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		processor := NewProcessor(l, r.Context(), db)

		vars := mux.Vars(r)
		id, err := uuid.Parse(vars["id"])
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid user ID", err.Error())
			return
		}

		model, err := processor.GetById(id)()
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				writeErrorResponse(w, http.StatusNotFound, "User not found", err.Error())
				return
			}
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to fetch user", err.Error())
			return
		}

		restModel, err := Transform(model)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to transform user", err.Error())
			return
		}

		writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": restModel})
	}
}

// createUserHandler handles POST /users
func createUserHandler(db *gorm.DB, l logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		processor := NewProcessor(l, r.Context(), db)

		var req CreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
			return
		}

		if req.Data.Type != "users" {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid resource type", "Expected type 'users'")
			return
		}

		input := CreateInput{
			Email:       req.Data.Attributes.Email,
			DisplayName: req.Data.Attributes.DisplayName,
			HouseholdId: req.Data.Attributes.HouseholdId,
		}

		model, err := processor.Create(input)()
		if err != nil {
			if errors.Is(err, ErrEmailAlreadyExists) {
				writeErrorResponse(w, http.StatusConflict, "Email already exists", err.Error())
				return
			}
			if errors.Is(err, ErrHouseholdNotFound) {
				writeErrorResponse(w, http.StatusBadRequest, "Household not found", err.Error())
				return
			}
			if errors.Is(err, ErrEmailRequired) || errors.Is(err, ErrEmailInvalid) ||
				errors.Is(err, ErrDisplayNameRequired) || errors.Is(err, ErrDisplayNameEmpty) {
				writeErrorResponse(w, http.StatusBadRequest, "Validation failed", err.Error())
				return
			}
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to create user", err.Error())
			return
		}

		restModel, err := Transform(model)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to transform user", err.Error())
			return
		}

		writeJSONResponse(w, http.StatusCreated, map[string]interface{}{"data": restModel})
	}
}

// updateUserHandler handles PATCH /users/:id
func updateUserHandler(db *gorm.DB, l logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		processor := NewProcessor(l, r.Context(), db)

		vars := mux.Vars(r)
		id, err := uuid.Parse(vars["id"])
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid user ID", err.Error())
			return
		}

		var req UpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
			return
		}

		if req.Data.Type != "users" {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid resource type", "Expected type 'users'")
			return
		}

		input := UpdateInput{
			Email:       req.Data.Attributes.Email,
			DisplayName: req.Data.Attributes.DisplayName,
		}

		model, err := processor.Update(id, input)()
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				writeErrorResponse(w, http.StatusNotFound, "User not found", err.Error())
				return
			}
			if errors.Is(err, ErrEmailAlreadyExists) {
				writeErrorResponse(w, http.StatusConflict, "Email already exists", err.Error())
				return
			}
			if errors.Is(err, ErrEmailRequired) || errors.Is(err, ErrEmailInvalid) ||
				errors.Is(err, ErrDisplayNameRequired) || errors.Is(err, ErrDisplayNameEmpty) {
				writeErrorResponse(w, http.StatusBadRequest, "Validation failed", err.Error())
				return
			}
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to update user", err.Error())
			return
		}

		restModel, err := Transform(model)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to transform user", err.Error())
			return
		}

		writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": restModel})
	}
}

// deleteUserHandler handles DELETE /users/:id
func deleteUserHandler(db *gorm.DB, l logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		processor := NewProcessor(l, r.Context(), db)

		vars := mux.Vars(r)
		id, err := uuid.Parse(vars["id"])
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid user ID", err.Error())
			return
		}

		err = processor.Delete(id)
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				writeErrorResponse(w, http.StatusNotFound, "User not found", err.Error())
				return
			}
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete user", err.Error())
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// associateHouseholdHandler handles POST /users/:id/relationships/household
func associateHouseholdHandler(db *gorm.DB, l logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		processor := NewProcessor(l, r.Context(), db)

		vars := mux.Vars(r)
		userId, err := uuid.Parse(vars["id"])
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid user ID", err.Error())
			return
		}

		var req AssociateHouseholdRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
			return
		}

		if req.Data.Type != "households" {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid resource type", "Expected type 'households'")
			return
		}

		householdId, err := uuid.Parse(req.Data.Id)
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid household ID", err.Error())
			return
		}

		model, err := processor.AssociateHousehold(userId, householdId)()
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				writeErrorResponse(w, http.StatusNotFound, "User not found", err.Error())
				return
			}
			if errors.Is(err, ErrHouseholdNotFound) {
				writeErrorResponse(w, http.StatusNotFound, "Household not found", err.Error())
				return
			}
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to associate household", err.Error())
			return
		}

		restModel, err := Transform(model)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to transform user", err.Error())
			return
		}

		writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": restModel})
	}
}

// disassociateHouseholdHandler handles DELETE /users/:id/relationships/household
func disassociateHouseholdHandler(db *gorm.DB, l logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		processor := NewProcessor(l, r.Context(), db)

		vars := mux.Vars(r)
		userId, err := uuid.Parse(vars["id"])
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid user ID", err.Error())
			return
		}

		model, err := processor.DisassociateHousehold(userId)()
		if err != nil {
			if errors.Is(err, ErrUserNotFound) {
				writeErrorResponse(w, http.StatusNotFound, "User not found", err.Error())
				return
			}
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to disassociate household", err.Error())
			return
		}

		restModel, err := Transform(model)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to transform user", err.Error())
			return
		}

		writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": restModel})
	}
}

// writeJSONResponse writes a JSON response with the given status code
func writeJSONResponse(w http.ResponseWriter, statusCode int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}

// writeErrorResponse writes a JSON:API error response
func writeErrorResponse(w http.ResponseWriter, statusCode int, title string, detail string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"errors": []map[string]interface{}{
			{
				"status": statusCode,
				"title":  title,
				"detail": detail,
			},
		},
	})
}
