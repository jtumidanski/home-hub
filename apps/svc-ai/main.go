package main

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/jtumidanski/home-hub/apps/svc-ai/handler"
	"github.com/jtumidanski/home-hub/apps/svc-ai/parser"
	"github.com/jtumidanski/home-hub/packages/shared-go/health"
	"github.com/jtumidanski/home-hub/packages/shared-go/logger"
	"github.com/jtumidanski/home-hub/packages/shared-go/rest/server"
	"github.com/jtumidanski/home-hub/packages/shared-go/service"
	"github.com/jtumidanski/home-hub/packages/shared-go/tracing"
	"github.com/sirupsen/logrus"
)

const (
	serviceName = "svc-ai"
	version     = "1.0.0"
)

type Server struct {
	baseUrl string
	prefix  string
}

func (s Server) GetBaseURL() string {
	return s.baseUrl
}

func (s Server) GetPrefix() string {
	return s.prefix
}

func GetServer() Server {
	return Server{
		baseUrl: "",
		prefix:  "/api/",
	}
}

func main() {
	l := logger.CreateLogger(serviceName)
	l.Infoln("Starting AI parsing service.")

	tdm := service.GetTeardownManager()

	tc, err := tracing.InitTracer(l)(serviceName)
	if err != nil {
		l.WithError(err).Fatal("Unable to initialize tracer.")
	}

	// Load configuration
	cfg, err := LoadConfig()
	if err != nil {
		l.WithError(err).Fatal("Failed to load configuration.")
	}
	l.WithField("provider", cfg.Provider).Info("Configuration loaded.")

	// Create ingredient parser (if enabled)
	var ingredientParser parser.IngredientParser
	if cfg.IsEnabled() {
		parserCfg := parser.Config{
			Provider: parser.ProviderType(cfg.Provider),
			Ollama: parser.OllamaConfig{
				BaseURL:   cfg.Ollama.BaseURL,
				ModelName: cfg.Ollama.ModelName,
				Timeout:   cfg.Ollama.Timeout,
			},
			Cloud: parser.CloudConfig{
				BaseURL:   cfg.Cloud.BaseURL,
				ModelName: cfg.Cloud.ModelName,
				APIKey:    cfg.Cloud.APIKey,
				Timeout:   cfg.Cloud.Timeout,
			},
		}
		ingredientParser, err = parser.NewParser(parserCfg, l)
		if err != nil {
			l.WithError(err).Fatal("Failed to create parser.")
		}
		l.WithField("parser", ingredientParser.Name()).Info("Parser created.")
	} else {
		l.Warn("AI parsing is disabled.")
	}

	// Create health check aggregator
	healthChecker := health.NewAggregator()
	// TODO: Add provider health check if enabled

	// Create route initializer
	routeInitializer := func(router *mux.Router, logger logrus.FieldLogger) {
		// Register health endpoint (unauthenticated)
		router.HandleFunc("/health", health.Handler(serviceName, version, healthChecker, logger)).Methods("GET")

		// Register parsing endpoints if parser is enabled
		if ingredientParser != nil {
			parseHandler := handler.NewParseHandler(ingredientParser, GetServer(), logger)
			router.HandleFunc("/parse/ingredient", parseHandler.HandleSingle).Methods("POST")
			router.HandleFunc("/parse/ingredients", parseHandler.HandleBatch).Methods("POST")
			router.HandleFunc("/parse/recipe", parseHandler.HandleRecipe).Methods("POST")
			logger.Info("Parsing endpoints registered.")
		} else {
			// Register placeholder endpoints that return error
			router.HandleFunc("/parse/ingredient", handleDisabled).Methods("POST")
			router.HandleFunc("/parse/ingredients", handleDisabled).Methods("POST")
			router.HandleFunc("/parse/recipe", handleDisabled).Methods("POST")
			logger.Warn("Parsing endpoints disabled.")
		}
	}

	// Create server with extended timeouts for AI processing
	server.New(l).
		WithContext(tdm.Context()).
		WithWaitGroup(tdm.WaitGroup()).
		SetBasePath(GetServer().GetPrefix()).
		SetReadTimeout(200 * time.Second).  // Long read timeout for large recipe text
		SetWriteTimeout(200 * time.Second). // Long write timeout for AI responses
		SetRouteInitializers(routeInitializer).
		Run()

	tdm.TeardownFunc(tracing.Teardown(l)(tc))

	tdm.Wait()
	l.Infoln("Service shutdown.")
}

// handleDisabled returns an error for disabled parsing endpoints
func handleDisabled(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   "provider_unavailable",
		"message": "AI provider is not configured or is disabled.",
	})
}
