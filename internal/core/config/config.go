// Package config reads process environment variables.
package config

import (
	"os"
	"strings"
)

// Config is the process configuration. Environment variables override defaults.
type Config struct {
	AppName     string
	Environment string
	LogLevel    string
	LogJSON     bool
	HTTPAddr    string
	DatabaseURL string
}

// Load reads configuration from the environment.
func Load() Config {
	logJSON := getenv("LOG_JSON", "false")
	return Config{
		AppName:     getenv("APP_NAME", "golang-template"),
		Environment: getenv("ENVIRONMENT", "local"),
		LogLevel:    getenv("LOG_LEVEL", "INFO"),
		LogJSON:     strings.EqualFold(logJSON, "true") || logJSON == "1",
		HTTPAddr:    getenv("HTTP_ADDR", ":8081"),
		DatabaseURL: getenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/app?sslmode=disable"),
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
