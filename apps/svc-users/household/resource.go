package household

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"

	userPkg "github.com/jtumidanski/home-hub/apps/svc-users/user"
	"github.com/jtumidanski/home-hub/packages/shared-go/rest/server"
)

// InitializeRoutes registers all household-related routes
func InitializeRoutes(db *gorm.DB) server.RouteInitializer {
	return func(router *mux.Router, l logrus.FieldLogger) {
		// Household CRUD endpoints
		router.HandleFunc("/households", listHouseholdsHandler(db, l)).Methods(http.MethodGet)
		router.HandleFunc("/households", createHouseholdHandler(db, l)).Methods(http.MethodPost)
		router.HandleFunc("/households/{id}", getHouseholdHandler(db, l)).Methods(http.MethodGet)
		router.HandleFunc("/households/{id}", updateHouseholdHandler(db, l)).Methods(http.MethodPatch)
		router.HandleFunc("/households/{id}", deleteHouseholdHandler(db, l)).Methods(http.MethodDelete)

		// Household relationships
		router.HandleFunc("/households/{id}/users", getHouseholdUsersHandler(db, l)).Methods(http.MethodGet)
	}
}

// listHouseholdsHandler handles GET /households
func listHouseholdsHandler(db *gorm.DB, l logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		models, err := GetAll(db)()
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to fetch households", err.Error())
			return
		}

		restModels, err := TransformSlice(models)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to transform households", err.Error())
			return
		}

		writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": restModels})
	}
}

// getHouseholdHandler handles GET /households/:id
func getHouseholdHandler(db *gorm.DB, l logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		processor := NewProcessor(l, r.Context(), db)

		vars := mux.Vars(r)
		id, err := uuid.Parse(vars["id"])
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid household ID", err.Error())
			return
		}

		model, err := processor.GetById(id)()
		if err != nil {
			if errors.Is(err, ErrHouseholdNotFound) {
				writeErrorResponse(w, http.StatusNotFound, "Household not found", err.Error())
				return
			}
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to fetch household", err.Error())
			return
		}

		restModel, err := Transform(model)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to transform household", err.Error())
			return
		}

		writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": restModel})
	}
}

// createHouseholdHandler handles POST /households
func createHouseholdHandler(db *gorm.DB, l logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		processor := NewProcessor(l, r.Context(), db)

		var req CreateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
			return
		}

		if req.Data.Type != "households" {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid resource type", "Expected type 'households'")
			return
		}

		input := CreateInput{
			Name: req.Data.Attributes.Name,
		}

		model, err := processor.Create(input)()
		if err != nil {
			if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrNameEmpty) {
				writeErrorResponse(w, http.StatusBadRequest, "Validation failed", err.Error())
				return
			}
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to create household", err.Error())
			return
		}

		restModel, err := Transform(model)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to transform household", err.Error())
			return
		}

		writeJSONResponse(w, http.StatusCreated, map[string]interface{}{"data": restModel})
	}
}

// updateHouseholdHandler handles PATCH /households/:id
func updateHouseholdHandler(db *gorm.DB, l logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		processor := NewProcessor(l, r.Context(), db)

		vars := mux.Vars(r)
		id, err := uuid.Parse(vars["id"])
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid household ID", err.Error())
			return
		}

		var req UpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
			return
		}

		if req.Data.Type != "households" {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid resource type", "Expected type 'households'")
			return
		}

		input := UpdateInput{
			Name: req.Data.Attributes.Name,
		}

		model, err := processor.Update(id, input)()
		if err != nil {
			if errors.Is(err, ErrHouseholdNotFound) {
				writeErrorResponse(w, http.StatusNotFound, "Household not found", err.Error())
				return
			}
			if errors.Is(err, ErrNameRequired) || errors.Is(err, ErrNameEmpty) {
				writeErrorResponse(w, http.StatusBadRequest, "Validation failed", err.Error())
				return
			}
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to update household", err.Error())
			return
		}

		restModel, err := Transform(model)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to transform household", err.Error())
			return
		}

		writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": restModel})
	}
}

// deleteHouseholdHandler handles DELETE /households/:id
func deleteHouseholdHandler(db *gorm.DB, l logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		processor := NewProcessor(l, r.Context(), db)

		vars := mux.Vars(r)
		id, err := uuid.Parse(vars["id"])
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid household ID", err.Error())
			return
		}

		err = processor.Delete(id)
		if err != nil {
			if errors.Is(err, ErrHouseholdNotFound) {
				writeErrorResponse(w, http.StatusNotFound, "Household not found", err.Error())
				return
			}
			if errors.Is(err, ErrHouseholdHasUsers) {
				writeErrorResponse(w, http.StatusConflict, "Cannot delete household with users", err.Error())
				return
			}
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to delete household", err.Error())
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// getHouseholdUsersHandler handles GET /households/:id/users
func getHouseholdUsersHandler(db *gorm.DB, l logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		processor := NewProcessor(l, r.Context(), db)

		vars := mux.Vars(r)
		id, err := uuid.Parse(vars["id"])
		if err != nil {
			writeErrorResponse(w, http.StatusBadRequest, "Invalid household ID", err.Error())
			return
		}

		// Verify household exists
		_, err = processor.GetById(id)()
		if err != nil {
			if errors.Is(err, ErrHouseholdNotFound) {
				writeErrorResponse(w, http.StatusNotFound, "Household not found", err.Error())
				return
			}
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to fetch household", err.Error())
			return
		}

		// Get users in household
		users, err := userPkg.GetByHouseholdId(db)(id)()
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to fetch users", err.Error())
			return
		}

		restModels, err := userPkg.TransformSlice(users)
		if err != nil {
			writeErrorResponse(w, http.StatusInternalServerError, "Failed to transform users", err.Error())
			return
		}

		writeJSONResponse(w, http.StatusOK, map[string]interface{}{"data": restModels})
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
