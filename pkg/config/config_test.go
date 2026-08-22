package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	// Clear envs safely
	_ = os.Unsetenv("LUMITREE_PORT")
	_ = os.Unsetenv("PORT")
	_ = os.Unsetenv("LUMITREE_HOST")
	_ = os.Unsetenv("LUMITREE_CACHE_TTL")
	_ = os.Unsetenv("LUMITREE_LOG_LEVEL")
	_ = os.Unsetenv("LUMITREE_LOG_FORMAT")
	_ = os.Unsetenv("LUMITREE_TIMETREE_BASE_URL")

	cfg := Load()

	if cfg.Port != 8080 {
		t.Errorf("expected Port=8080, got %d", cfg.Port)
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("expected Host='0.0.0.0', got '%s'", cfg.Host)
	}
	if cfg.CacheTTL != 10*time.Minute {
		t.Errorf("expected CacheTTL=10m, got %v", cfg.CacheTTL)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected LogLevel='info', got '%s'", cfg.LogLevel)
	}
	if cfg.LogFormat != "text" {
		t.Errorf("expected LogFormat='text', got '%s'", cfg.LogFormat)
	}
	if cfg.TimeTreeBaseURL != "https://timetreeapp.com" {
		t.Errorf("expected TimeTreeBaseURL='https://timetreeapp.com', got '%s'", cfg.TimeTreeBaseURL)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("LUMITREE_PORT", "9090")
	t.Setenv("LUMITREE_HOST", "127.0.0.1")
	t.Setenv("LUMITREE_CACHE_TTL", "30m")
	t.Setenv("LUMITREE_LOG_LEVEL", "debug")
	t.Setenv("LUMITREE_LOG_FORMAT", "json")
	t.Setenv("LUMITREE_TIMETREE_BASE_URL", "http://mock-server:8000")

	cfg := Load()

	if cfg.Port != 9090 {
		t.Errorf("expected Port=9090, got %d", cfg.Port)
	}
	if cfg.Host != "127.0.0.1" {
		t.Errorf("expected Host='127.0.0.1', got '%s'", cfg.Host)
	}
	if cfg.CacheTTL != 30*time.Minute {
		t.Errorf("expected CacheTTL=30m, got %v", cfg.CacheTTL)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected LogLevel='debug', got '%s'", cfg.LogLevel)
	}
	if cfg.LogFormat != "json" {
		t.Errorf("expected LogFormat='json', got '%s'", cfg.LogFormat)
	}
	if cfg.TimeTreeBaseURL != "http://mock-server:8000" {
		t.Errorf("expected TimeTreeBaseURL='http://mock-server:8000', got '%s'", cfg.TimeTreeBaseURL)
	}
}
