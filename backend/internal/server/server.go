package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/logiflows/logiflows/backend/internal/config"
)

// Server wraps the http.Server for LogiFlows.
type Server struct {
	httpServer *http.Server
	log        *slog.Logger
	cfg        *config.Config
}

// New creates a new HTTP Server instance.
func New(cfg *config.Config, log *slog.Logger, router *gin.Engine) *Server {
	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)

	return &Server{
		httpServer: &http.Server{
			Addr:         addr,
			Handler:      router,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		log: log,
		cfg: cfg,
	}
}

// Start runs the HTTP server listening on the configured address in a non-blocking goroutine.
func (s *Server) Start() <-chan error {
	errChan := make(chan error, 1)

	go func() {
		s.log.Info("Starting HTTP server",
			"service", s.cfg.App.Name,
			"env", s.cfg.App.Env,
			"addr", s.httpServer.Addr,
		)

		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("HTTP server ListenAndServe error: %w", err)
			return
		}
		close(errChan)
	}()

	return errChan
}

// Shutdown gracefully stops accepting new connections and flushes active requests.
func (s *Server) Shutdown(ctx context.Context) error {
	s.log.Info("Initiating graceful HTTP server shutdown...")
	return s.httpServer.Shutdown(ctx)
}
