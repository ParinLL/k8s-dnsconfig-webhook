package main

import (
	"log"

	"github.com/ParinLL/k8s-dnsconfig-webhook/internal/config"
	"github.com/ParinLL/k8s-dnsconfig-webhook/internal/server"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create and start the webhook server
	srv := server.New(cfg)
	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
