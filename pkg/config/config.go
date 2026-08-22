// Package config provides configuration management for lumitree using environment variables.
package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration parameters for lumitree.
type Config struct {
	Port            int           `json:"port"`
	Host            string        `json:"host"`
	CacheTTL        time.Duration `json:"cache_ttl"`
	LogLevel        string        `json:"log_level"`
	LogFormat       string        `json:"log_format"`
	TimeTreeBaseURL string        `json:"timetree_base_url"`
}

// Load loads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		Port:            getEnvAsInt("LUMITREE_PORT", getEnvAsInt("PORT", 8080)),
		Host:            getEnv("LUMITREE_HOST", getEnv("HOST", "0.0.0.0")),
		CacheTTL:        getEnvAsDuration("LUMITREE_CACHE_TTL", 10*time.Minute),
		LogLevel:        getEnv("LUMITREE_LOG_LEVEL", getEnv("LOG_LEVEL", "info")),
		LogFormat:       getEnv("LUMITREE_LOG_FORMAT", getEnv("LOG_FORMAT", "text")),
		TimeTreeBaseURL: getEnv("LUMITREE_TIMETREE_BASE_URL", "https://timetreeapp.com"),
	}
}

func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok && val != "" {
		return val
	}
	return defaultVal
}

func getEnvAsInt(key string, defaultVal int) int {
	if valStr, ok := os.LookupEnv(key); ok && valStr != "" {
		if val, err := strconv.Atoi(valStr); err == nil {
			return val
		}
	}
	return defaultVal
}

func getEnvAsDuration(key string, defaultVal time.Duration) time.Duration {
	if valStr, ok := os.LookupEnv(key); ok && valStr != "" {
		if val, err := time.ParseDuration(valStr); err == nil {
			return val
		}
	}
	return defaultVal
}
