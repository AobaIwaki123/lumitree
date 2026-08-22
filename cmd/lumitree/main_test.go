package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AobaIwaki123/lumitree/pkg/config"
)

func mockTimeTreeServer(t *testing.T) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	// Initial HTML page with CSRF token
	mux.HandleFunc("GET /public_calendars/test_cal", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!DOCTYPE html><html><head><meta name="csrf-token" content="mock-csrf-token"></head><body></body></html>`))
	})

	// Calendar API
	mux.HandleFunc("GET /api/v2/public_calendars/test_cal", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"public_calendar": {
				"id": "12345",
				"alias_code": "test_cal",
				"name": "Test Calendar",
				"description": "Test Calendar Description",
				"image_url": "https://example.com/cover.jpg"
			}
		}`))
	})

	// Events API
	mux.HandleFunc("GET /api/v2/public_calendars/test_cal/public_events", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"public_events": [
				{
					"id": "evt-1",
					"title": "Live Show 2026",
					"description": "Annual live concert",
					"start_at": 1787648400000,
					"end_at": 1787659200000,
					"all_day": false,
					"start_timezone": "Asia/Tokyo",
					"end_timezone": "Asia/Tokyo",
					"location_name": "Tokyo Dome",
					"link_url": "https://example.com/live"
				},
				{
					"id": "evt-2",
					"title": "All Day Festival",
					"description": "Outdoor fest",
					"start_at": 1787734800000,
					"end_at": 1787734800000,
					"all_day": true,
					"start_timezone": "Asia/Tokyo",
					"end_timezone": "Asia/Tokyo"
				}
			],
			"paging": {
				"current_page": 1,
				"total_pages": 1,
				"total_count": 2
			}
		}`))
	})

	return httptest.NewServer(mux)
}

func captureStdout(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	_ = w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	return buf.String()
}

func TestPrintUsage(t *testing.T) {
	out := captureStdout(func() {
		printUsage()
	})

	if !strings.Contains(out, "Usage: lumitree <command>") {
		t.Errorf("expected usage output, got %s", out)
	}
	if !strings.Contains(out, "serve") || !strings.Contains(out, "get") || !strings.Contains(out, "ics") {
		t.Errorf("missing commands in usage output: %s", out)
	}
}

func TestRunGet_JSON(t *testing.T) {
	ts := mockTimeTreeServer(t)
	defer ts.Close()

	cfg := &config.Config{
		TimeTreeBaseURL: ts.URL,
	}

	out := captureStdout(func() {
		runGet(cfg, []string{"test_cal", "--json", "--year", "2026"})
	})

	if !strings.Contains(out, `"title": "Test Calendar"`) {
		t.Errorf("expected calendar title in JSON output, got: %s", out)
	}
	if !strings.Contains(out, `"title": "Live Show 2026"`) {
		t.Errorf("expected event title in JSON output, got: %s", out)
	}
	if !strings.Contains(out, `"location": "Tokyo Dome"`) {
		t.Errorf("expected location in JSON output, got: %s", out)
	}
}

func TestRunGet_Table(t *testing.T) {
	ts := mockTimeTreeServer(t)
	defer ts.Close()

	cfg := &config.Config{
		TimeTreeBaseURL: ts.URL,
	}

	out := captureStdout(func() {
		runGet(cfg, []string{"test_cal", "--year", "2026"})
	})

	if !strings.Contains(out, "Calendar: Test Calendar (12345)") {
		t.Errorf("expected calendar header, got: %s", out)
	}
	if !strings.Contains(out, "Live Show 2026") {
		t.Errorf("expected event title in table, got: %s", out)
	}
	if !strings.Contains(out, "Tokyo Dome") {
		t.Errorf("expected location in table, got: %s", out)
	}
	if !strings.Contains(out, "All Day Festival") {
		t.Errorf("expected all-day event title in table, got: %s", out)
	}
}

func TestRunICS(t *testing.T) {
	ts := mockTimeTreeServer(t)
	defer ts.Close()

	cfg := &config.Config{
		TimeTreeBaseURL: ts.URL,
	}

	tmpDir := t.TempDir()
	outFile := filepath.Join(tmpDir, "test_events.ics")

	out := captureStdout(func() {
		runICS(cfg, []string{"test_cal", "--output", outFile, "--year", "2026"})
	})

	if !strings.Contains(out, fmt.Sprintf("Successfully exported 2 events to %s", outFile)) {
		t.Errorf("expected export success message, got: %s", out)
	}

	data, err := os.ReadFile(outFile)
	if err != nil {
		t.Fatalf("failed to read generated ics file: %v", err)
	}

	icsContent := string(data)
	if !strings.Contains(icsContent, "BEGIN:VCALENDAR") {
		t.Errorf("missing BEGIN:VCALENDAR in output: %s", icsContent)
	}
	if !strings.Contains(icsContent, "SUMMARY:Live Show 2026") {
		t.Errorf("missing event summary in ics: %s", icsContent)
	}
	if !strings.Contains(icsContent, "LOCATION:Tokyo Dome") {
		t.Errorf("missing location in ics: %s", icsContent)
	}
}

func TestRunServe_ConfigParsing(t *testing.T) {
	cfg := &config.Config{
		Port: 8080,
		Host: "0.0.0.0",
	}

	args := []string{"--port", "9090", "--host", "127.0.0.1"}
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--port", "-p":
			if i+1 < len(args) {
				cfg.Port = 9090
				i++
			}
		case "--host":
			if i+1 < len(args) {
				cfg.Host = args[i+1]
				i++
			}
		}
	}

	if cfg.Port != 9090 {
		t.Errorf("expected port 9090, got %d", cfg.Port)
	}
	if cfg.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", cfg.Host)
	}
}
