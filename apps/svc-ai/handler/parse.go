package handler

import (
	"encoding/json"
	"net/http"

	"github.com/jtumidanski/api2go/jsonapi"
	"github.com/jtumidanski/home-hub/apps/svc-ai/parser"
	"github.com/jtumidanski/home-hub/packages/shared-go/rest/server"
	"github.com/sirupsen/logrus"
)

// ParseHandler handles ingredient parsing requests
type ParseHandler struct {
	parser     parser.IngredientParser
	logger     logrus.FieldLogger
	serverInfo jsonapi.ServerInformation
}

// NewParseHandler creates a new parse handler
func NewParseHandler(p parser.IngredientParser, si jsonapi.ServerInformation, logger logrus.FieldLogger) *ParseHandler {
	return &ParseHandler{
		parser:     p,
		serverInfo: si,
		logger:     logger,
	}
}

// SingleParseRequest represents a request to parse a single ingredient line
type SingleParseRequest struct {
	Line   string             `json:"line"`
	Locale string             `json:"locale,omitempty"`
	Hints  *parser.ParseHints `json:"hints,omitempty"`
}

// BatchParseRequest represents a request to parse multiple ingredient lines
type BatchParseRequest struct {
	Lines  []string           `json:"lines"`
	Locale string             `json:"locale,omitempty"`
	Hints  *parser.ParseHints `json:"hints,omitempty"`
}

// RecipeParseRequest represents a request to parse a full recipe
type RecipeParseRequest struct {
	RecipeText string             `json:"recipeText"`
	Locale     string             `json:"locale,omitempty"`
	Hints      *parser.ParseHints `json:"hints,omitempty"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// HandleSingle handles POST /parse/ingredient
func (h *ParseHandler) HandleSingle(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req SingleParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request")
		h.writeError(w, "invalid_request", "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if req.Line == "" {
		h.writeError(w, "invalid_request", "Line field is required", http.StatusBadRequest)
		return
	}

	// Build parse options
	opts := parser.ParseOptions{
		Locale: req.Locale,
	}
	if req.Hints != nil {
		opts.Hints = *req.Hints
	}

	// Parse ingredient
	result, err := h.parser.Parse(r.Context(), req.Line, opts)
	if err != nil {
		h.logger.WithError(err).WithField("line", req.Line).Error("Failed to parse ingredient")
		h.writeError(w, "parsing_failed", err.Error(), http.StatusInternalServerError)
		return
	}

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(result)
}

// HandleBatch handles POST /parse/ingredients
func (h *ParseHandler) HandleBatch(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req BatchParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request")
		h.writeError(w, "invalid_request", "Failed to parse request body", http.StatusBadRequest)
		return
	}

	// Validate request
	if len(req.Lines) == 0 {
		h.writeError(w, "invalid_request", "Lines field is required and must not be empty", http.StatusBadRequest)
		return
	}

	// Enforce max batch size
	maxBatchSize := 50
	if len(req.Lines) > maxBatchSize {
		h.writeError(w, "invalid_request", "Maximum batch size is 50 lines", http.StatusBadRequest)
		return
	}

	// Build parse options
	opts := parser.ParseOptions{
		Locale: req.Locale,
	}
	if req.Hints != nil {
		opts.Hints = *req.Hints
	}

	// Parse ingredients
	results, err := h.parser.ParseBatch(r.Context(), req.Lines, opts)
	if err != nil {
		h.logger.WithError(err).WithField("line_count", len(req.Lines)).Error("Failed to parse ingredients")
		h.writeError(w, "parsing_failed", err.Error(), http.StatusInternalServerError)
		return
	}

	// Write response
	response := parser.BatchParseResult{
		Results: results,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HandleRecipe handles POST /parse/recipe
func (h *ParseHandler) HandleRecipe(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req RecipeParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.WithError(err).Error("Failed to decode request")
		h.writeError(w, "invalid_request", "Failed to parse request body", http.StatusBadRequest)
		return
	}

	h.logger.WithField("recipe_text_length", len(req.RecipeText)).Debug("Starting recipe parsing")

	// Validate request
	if req.RecipeText == "" {
		h.writeError(w, "invalid_request", "RecipeText field is required", http.StatusBadRequest)
		return
	}

	// Build parse options
	opts := parser.ParseOptions{
		Locale: req.Locale,
	}
	if req.Hints != nil {
		opts.Hints = *req.Hints
	}

	// Parse recipe
	h.logger.Debug("Calling parser.ParseRecipe")
	results, err := h.parser.ParseRecipe(r.Context(), req.RecipeText, opts)
	if err != nil {
		h.logger.WithError(err).Error("Failed to parse recipe")
		h.writeError(w, "parsing_failed", err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.WithField("ingredient_count", len(results)).Debug("Recipe parsing completed successfully")

	// Transform to REST models
	h.logger.Debug("Transforming results to REST models")
	restModels, err := parser.TransformParseResultSlice(results)
	if err != nil {
		h.logger.WithError(err).Error("Failed to transform results to REST models")
		h.writeError(w, "transformation_failed", err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.WithField("rest_model_count", len(restModels)).Debug("REST models created")

	// Marshal to check size (for debugging)
	testBytes, _ := json.Marshal(restModels)
	h.logger.WithField("response_size_bytes", len(testBytes)).Debug("Sending JSON:API response to client")

	// Use server.MarshalResponse for JSON:API format
	server.MarshalResponse[[]parser.ParsedIngredientRestModel](h.logger)(w)(h.serverInfo)(map[string][]string{})(restModels)

	h.logger.Debug("JSON:API response sent successfully")
}

// writeError writes a JSON error response
func (h *ParseHandler) writeError(w http.ResponseWriter, errorCode, message string, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   errorCode,
		Message: message,
	})
}
