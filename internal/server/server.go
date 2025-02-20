package server

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/ParinLL/k8s-dnsconfig-webhook/internal/admission"
	"github.com/ParinLL/k8s-dnsconfig-webhook/internal/config"
	"k8s.io/klog/v2"
)

func init() {
	klog.InitFlags(nil)
	// Get log level from environment variable
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel != "" {
		flag.Set("v", logLevel)
		flag.Parse()
	}
}

// Server represents the webhook server
type Server struct {
	config *config.Config
	mux    *http.ServeMux
}

// New creates a new webhook server
func New(cfg *config.Config) *Server {
	klog.V(2).Info("Initializing webhook server")

	server := &Server{
		config: cfg,
		mux:    http.NewServeMux(),
	}

	// Create the mutator and handler
	configData, err := server.loadDNSConfig()
	if err != nil {
		klog.Fatalf("Failed to load DNS config: %v", err)
	}

	mutator, err := admission.NewDNSConfigMutator(configData)
	if err != nil {
		klog.Fatalf("Failed to create mutator: %v", err)
	}

	handler := admission.NewHandler(mutator)

	// Start ConfigMap watcher
	go server.watchConfigMap(mutator)

	// Register routes
	server.mux.Handle("/mutate", handler)
	server.mux.HandleFunc("/health", server.healthCheck)

	klog.V(2).Info("Server initialization completed")
	return server
}

// Start begins serving webhook requests
func (s *Server) Start() error {
	startTime := time.Now()
	addr := fmt.Sprintf(":%d", s.config.Port)
	klog.Info("Starting webhook server on", addr)

	// Check if cert files exist and are readable
	if _, err := os.Stat(s.config.CertFile); err != nil {
		klog.Errorf("Certificate file issue: %v", err)
		return err
	}
	if _, err := os.Stat(s.config.KeyFile); err != nil {
		klog.Errorf("Key file issue: %v", err)
		return err
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           s.logRequests(s.mux),
		ReadHeaderTimeout: 3 * time.Second,
	}

	// Log startup time
	klog.V(2).Infof("Server initialization took %v", time.Since(startTime))

	return server.ListenAndServeTLS(s.config.CertFile, s.config.KeyFile)
}

// healthCheck handles health check requests
func (s *Server) healthCheck(w http.ResponseWriter, r *http.Request) {
	if klog.V(4).Enabled() {
		klog.V(4).Infof("Health check from %s", r.RemoteAddr)
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// loadDNSConfig loads the DNS configuration from the ConfigMap
func (s *Server) loadDNSConfig() ([]byte, error) {
	configPath := "/etc/webhook/config/config.yaml"

	// Read from ConfigMap
	data, err := os.ReadFile(configPath)
	if err != nil {
		klog.Errorf("Failed to read DNS config from ConfigMap: %v", err)
		// Return default configuration if ConfigMap read fails
		return []byte(`
dnsConfig:
  options:
    - name: ndots
      value: "1"
`), nil
	}

	klog.V(2).Infof("Loaded DNS configuration from ConfigMap: %s", string(data))
	return data, nil
}

// logRequests is a middleware that logs all incoming requests
func (s *Server) logRequests(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only log non-health check requests at V(2)
		if r.URL.Path != "/health" {
			klog.V(2).Infof("Received request: method=%s, path=%s, remote=%s",
				r.Method, r.URL.Path, r.RemoteAddr)
		}
		handler.ServeHTTP(w, r)
	})
}

// watchConfigMap monitors the ConfigMap for changes
func (s *Server) watchConfigMap(mutator *admission.DNSConfigMutator) {
	configPath := "/etc/webhook/config/config.yaml"
	lastMod := time.Now()

	for {
		time.Sleep(1 * time.Second)

		stat, err := os.Stat(configPath)
		if err != nil {
			klog.Errorf("Failed to stat config file: %v", err)
			continue
		}

		// Check if file has been modified
		if stat.ModTime().After(lastMod) {
			klog.V(2).Info("ConfigMap change detected, reloading configuration")

			data, err := os.ReadFile(configPath)
			if err != nil {
				klog.Errorf("Failed to read updated config: %v", err)
				continue
			}

			if err := mutator.UpdateConfig(data); err != nil {
				klog.Errorf("Failed to update DNS configuration: %v", err)
				continue
			}

			lastMod = stat.ModTime()
			klog.V(2).Info("Successfully updated DNS configuration from ConfigMap")
		}
	}
}
