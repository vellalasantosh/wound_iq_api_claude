package config

import (
	"fmt"
	"os"
)

// Config holds application configuration
type Config struct {
	DBDSN string
	Port  string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	dbDSN := os.Getenv("DB_DSN")
	if dbDSN == "" {
		return nil, fmt.Errorf("DB_DSN environment variable is required")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // default port
	}

	return &Config{
		DBDSN: dbDSN,
		Port:  port,
	}, nil
}
