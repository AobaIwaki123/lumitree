package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AobaIwaki123/lumitree/pkg/config"
	"github.com/AobaIwaki123/lumitree/pkg/timetree"
)

func TestServerEndpoints(t *testing.T) {
	pageHTML, _ := os.ReadFile(filepath.Join("..", "..", "testdata", "sample_page.html"))
	calJSON, _ := os.ReadFile(filepath.Join("..", "..", "testdata", "sample_calendar.json"))
	eventsJSON, _ := os.ReadFile(filepath.Join("..", "..", "testdata", "sample_events.json"))

	mockTimeTree := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/public_calendars/ilife_official":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(pageHTML)
		case "/api/v2/public_calendars/ilife_official":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(calJSON)
		case "/api/v2/public_calendars/ilife_official/public_events":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(eventsJSON)
		default:
			http.NotFound(w, r)
		}
	}))
	defer mockTimeTree.Close()

	cfg := &config.Config{
		CacheTTL:        5 * time.Minute,
		TimeTreeBaseURL: mockTimeTree.URL,
	}

	client, err := timetree.NewClient(timetree.WithBaseURL(mockTimeTree.URL))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	srv := NewServer(cfg, client, nil)
	handler := srv.Handler()

	// 1. Test /healthz
	t.Run("GET /healthz", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"status": "ok"`) {
			t.Errorf("expected status ok, got %s", rec.Body.String())
		}
	})

	// 2. Test /api/v1/calendars/ilife_official
	t.Run("GET /api/v1/calendars/ilife_official", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/calendars/ilife_official", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), `"title": "iLiFE!"`) {
			t.Errorf("expected iLiFE!, got %s", rec.Body.String())
		}
	})

	// 3. Test /api/v1/calendars/ilife_official/events
	t.Run("GET /api/v1/calendars/ilife_official/events", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/calendars/ilife_official/events", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		if !strings.Contains(rec.Body.String(), "MEGALiFE!") {
			t.Errorf("expected MEGALiFE!, got %s", rec.Body.String())
		}
	})

	// 4. Test /api/v1/calendars/ilife_official/events.ics
	t.Run("GET /api/v1/calendars/ilife_official/events.ics", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/calendars/ilife_official/events.ics", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", rec.Code)
		}
		if rec.Header().Get("Content-Type") != "text/calendar; charset=utf-8" {
			t.Errorf("expected text/calendar content-type, got %s", rec.Header().Get("Content-Type"))
		}
		if !strings.Contains(rec.Body.String(), "BEGIN:VCALENDAR") {
			t.Errorf("expected VCALENDAR, got %s", rec.Body.String())
		}
	})
}
