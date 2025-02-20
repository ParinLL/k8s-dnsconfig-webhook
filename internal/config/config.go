package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds the server configuration
type Config struct {
	Port     int
	CertFile string
	KeyFile  string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	port := getEnvInt("WEBHOOK_PORT", 8443)
	certFile := getEnvString("WEBHOOK_CERT_FILE", "/etc/webhook/certs/tls.crt")
	keyFile := getEnvString("WEBHOOK_KEY_FILE", "/etc/webhook/certs/tls.key")

	// Validate certificate files exist
	if _, err := os.Stat(certFile); err != nil {
		return nil, fmt.Errorf("cert file not found: %v", err)
	}
	if _, err := os.Stat(keyFile); err != nil {
		return nil, fmt.Errorf("key file not found: %v", err)
	}

	return &Config{
		Port:     port,
		CertFile: certFile,
		KeyFile:  keyFile,
	}, nil
}

// Helper functions for environment variables
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}
