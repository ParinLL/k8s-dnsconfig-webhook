package server

import (
	"fmt"
	"net/http"

	"github.com/your-org/dns-webhook/internal/admission"
	"github.com/your-org/dns-webhook/internal/config"
	"k8s.io/klog/v2"
)

// Server represents the webhook server
type Server struct {
	config *config.Config
	mux    *http.ServeMux
}

// New creates a new webhook server
func New(cfg *config.Config) *Server {
	server := &Server{
		config: cfg,
		mux:    http.NewServeMux(),
	}

	// Create the mutator and handler
	mutator := admission.NewDNSConfigMutator("1")
	handler := admission.NewHandler(mutator)

	// Register routes
	server.mux.Handle("/mutate", handler)
	server.mux.HandleFunc("/health", server.healthCheck)

	return server
}

// Start begins serving webhook requests
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.config.Port)
	klog.Infof("Starting webhook server on %s", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: s.mux,
	}

	return server.ListenAndServeTLS(s.config.CertFile, s.config.KeyFile)
}

// healthCheck handles health check requests
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}
