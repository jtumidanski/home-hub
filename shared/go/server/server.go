// Package server provides the HTTP server with middleware, health endpoints,
// and JSON:API handler registration.
package server

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

// RouteInitializer is a function that registers routes on a router.
type RouteInitializer func(router *mux.Router)

// Server is the HTTP server builder.
type Server struct {
	logger       *logrus.Logger
	router       *mux.Router
	initializers []RouteInitializer
	addr         string
}

// New creates a new Server with default configuration.
func New(logger *logrus.Logger) *Server {
	r := mux.NewRouter()
	return &Server{
		logger: logger,
		router: r,
		addr:   ":8080",
	}
}

// WithAddr sets the listen address.
func (s *Server) WithAddr(addr string) *Server {
	s.addr = addr
	return s
}

// AddRouteInitializer adds a route initializer to be called before the server starts.
func (s *Server) AddRouteInitializer(ri RouteInitializer) *Server {
	s.initializers = append(s.initializers, ri)
	return s
}

// Run starts the server with graceful shutdown support.
func (s *Server) Run() {
	s.router.Use(RequestIDMiddleware)
	s.router.Use(TracingMiddleware)
	s.router.Use(LoggingMiddleware(s.logger))

	s.router.HandleFunc("/healthz", HealthHandler).Methods(http.MethodGet)
	s.router.HandleFunc("/readyz", ReadyHandler).Methods(http.MethodGet)

	for _, init := range s.initializers {
		init(s.router)
	}

	srv := &http.Server{
		Addr:         s.addr,
		Handler:      s.router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		s.logger.WithField("addr", s.addr).Info("server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.WithError(err).Fatal("server failed")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	s.logger.Info("server shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		s.logger.WithError(err).Fatal("server forced shutdown")
	}
	s.logger.Info("server stopped")
}
