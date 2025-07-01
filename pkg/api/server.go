package api

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/config"
	"github.com/AkshayDubey29/mimir-limit-optimizer/internal/controller"
)

// Server represents the API server for the web UI
type Server struct {
	controller *controller.MimirLimitController
	config     *config.Config
	log        logr.Logger
	router     *mux.Router
	httpServer *http.Server
	uiAssets   embed.FS
}

// NewServer creates a new API server instance
func NewServer(controller *controller.MimirLimitController, cfg *config.Config, log logr.Logger, uiAssets embed.FS) *Server {
	s := &Server{
		controller: controller,
		config:     cfg,
		log:        log,
		router:     mux.NewRouter(),
		uiAssets:   uiAssets,
	}
	
	s.setupRoutes()
	return s
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Add middleware
	s.router.Use(s.loggingMiddleware)
	s.router.Use(s.corsMiddleware)
	
	// API routes
	api := s.router.PathPrefix("/api").Subrouter()
	
	// System endpoints
	api.HandleFunc("/status", s.handleStatus).Methods("GET")
	api.HandleFunc("/config", s.handleConfig).Methods("GET", "POST")
	api.HandleFunc("/metrics", s.handleMetrics).Methods("GET")
	
	// Tenant endpoints
	api.HandleFunc("/tenants", s.handleTenants).Methods("GET")
	api.HandleFunc("/tenants/{tenant_id}", s.handleTenantDetail).Methods("GET")
	
	// Analysis endpoints
	api.HandleFunc("/diff", s.handleDiff).Methods("GET")
	api.HandleFunc("/audit", s.handleAudit).Methods("GET")
	
	// Test endpoints
	api.HandleFunc("/test/spike", s.handleTestSpike).Methods("POST")
	api.HandleFunc("/test/alert", s.handleTestAlert).Methods("POST")
	api.HandleFunc("/test/reconcile", s.handleTestReconcile).Methods("POST")
	
	// Health check
	s.router.HandleFunc("/health", s.handleHealthCheck).Methods("GET")
	
	// Prometheus metrics endpoint
	s.router.Handle("/metrics", promhttp.Handler())
	
	// Setup UI static file serving - embed.FS is always valid, so check if we can access the UI directory
	if _, err := s.uiAssets.Open("ui"); err == nil {
		s.setupUIRoutes()
	}
}

// setupUIRoutes configures UI static file serving
func (s *Server) setupUIRoutes() {
	// Setup static UI file server
	uiBuildFS, err := fs.Sub(s.uiAssets, "ui/build")
	if err != nil {
		s.log.Error(err, "failed to create UI filesystem")
		return
	}
	
	// Serve static files (CSS, JS, images, etc.)
	staticFS, err := fs.Sub(uiBuildFS, "static")
	if err == nil {
		s.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))
	}
	
	// Serve favicon and other root assets
	s.router.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		if file, err := uiBuildFS.Open("favicon.ico"); err == nil {
			defer file.Close()
			http.ServeContent(w, r, "favicon.ico", time.Time{}, file.(io.ReadSeeker))
		} else {
			http.NotFound(w, r)
		}
	})
	
	// Serve the React app for all non-API routes (SPA fallback)
	s.router.PathPrefix("/").HandlerFunc(s.serveReactApp(uiBuildFS))
}

// serveReactApp serves the React application for SPA routing
func (s *Server) serveReactApp(uiBuildFS fs.FS) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the requested file
		filePath := r.URL.Path
		if filePath == "/" {
			filePath = "/index.html"
		}
		
		// Remove leading slash for fs.FS
		if len(filePath) > 0 && filePath[0] == '/' {
			filePath = filePath[1:]
		}
		
		// Check if file exists
		if file, err := uiBuildFS.Open(filePath); err == nil {
			file.Close()
			// File exists, serve it
			http.FileServer(http.FS(uiBuildFS)).ServeHTTP(w, r)
			return
		}
		
		// File doesn't exist, serve index.html for SPA routing
		indexFile, err := uiBuildFS.Open("index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer indexFile.Close()
		
		// Set content type for HTML
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		
		// Copy index.html content to response
		http.ServeContent(w, r, "index.html", time.Time{}, indexFile.(io.ReadSeeker))
	}
}

// Start starts the HTTP server
func (s *Server) Start(port int) error {
	addr := fmt.Sprintf(":%d", port)
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	
	s.log.Info("starting API server", "port", port)
	return s.httpServer.ListenAndServe()
}

// Stop gracefully stops the HTTP server
func (s *Server) Stop(ctx context.Context) error {
	if s.httpServer != nil {
		return s.httpServer.Shutdown(ctx)
	}
	return nil
}

// Middleware
func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		duration := time.Since(start)
		
		s.log.V(1).Info("API request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration", duration,
			"remote_addr", r.RemoteAddr,
		)
	})
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		
		next.ServeHTTP(w, r)
	})
}

// Error handling
func (s *Server) writeError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"error":     true,
		"message":   message,
		"timestamp": time.Now().Format(time.RFC3339),
	}); err != nil {
		s.log.Error(err, "failed to encode error response")
		// Note: Cannot modify headers or write additional content after WriteHeader() is called
	}
}

func (s *Server) writeJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		s.log.Error(err, "failed to encode JSON response")
		s.writeError(w, http.StatusInternalServerError, "Failed to encode response")
	}
} 