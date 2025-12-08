// Package webhook provides HTTP server infrastructure for webhooks.
package webhook

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/exernia/botjanweb/pkg/logger"
)

// Server wraps an HTTP server for webhook endpoints.
type Server struct {
	httpServer *http.Server
	logger     *log.Logger
}

// NewServer creates a new webhook server.
//
// Parameters:
//   - port: Port number to listen on
//   - handler: HTTP handler for processing requests
func NewServer(port int, handler http.Handler) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf(":%d", port),
			Handler:      handler,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		logger: logger.Webhook,
	}
}

// Start begins listening for incoming webhook requests.
// This is non-blocking and runs the server in a goroutine.
func (s *Server) Start() error {
	s.logger.Printf("Server starting on %s", s.httpServer.Addr)

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Printf("Server error: %v", err)
		}
	}()

	return nil
}

// Stop gracefully shuts down the webhook server.
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Println("Shutting down server...")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	s.logger.Println("Server stopped")
	return nil
}

// Addr returns the server address.
func (s *Server) Addr() string {
	return s.httpServer.Addr
}
