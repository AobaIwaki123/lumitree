package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/AobaIwaki123/lumitree/pkg/config"
	"github.com/AobaIwaki123/lumitree/pkg/server"
	"github.com/AobaIwaki123/lumitree/pkg/timetree"
)

func TestGeneratedClientWithServer(t *testing.T) {
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
		TimeTreeBaseURL: mockTimeTree.URL,
	}

	ttClient, err := timetree.NewClient(timetree.WithBaseURL(mockTimeTree.URL))
	if err != nil {
		t.Fatalf("failed to create TimeTree client: %v", err)
	}

	srv := server.NewServer(cfg, ttClient, nil)
	testServer := httptest.NewServer(srv.Handler())
	defer testServer.Close()

	// Use generated ClientWithResponses to call the server
	client, err := NewClientWithResponses(testServer.URL)
	if err != nil {
		t.Fatalf("failed to create generated API client: %v", err)
	}

	ctx := context.Background()

	// 1. Test GetHealth
	healthResp, err := client.GetHealthWithResponse(ctx)
	if err != nil {
		t.Fatalf("GetHealth failed: %v", err)
	}
	if healthResp.StatusCode() != http.StatusOK {
		t.Errorf("expected 200, got %d", healthResp.StatusCode())
	}
	if healthResp.JSON200.Status != "ok" {
		t.Errorf("expected status 'ok', got '%s'", healthResp.JSON200.Status)
	}

	// 2. Test GetCalendar
	calResp, err := client.GetCalendarWithResponse(ctx, "ilife_official")
	if err != nil {
		t.Fatalf("GetCalendar failed: %v", err)
	}
	if calResp.StatusCode() != http.StatusOK {
		t.Errorf("expected 200, got %d", calResp.StatusCode())
	}
	if calResp.JSON200.Title != "iLiFE!" {
		t.Errorf("expected title 'iLiFE!', got '%s'", calResp.JSON200.Title)
	}

	// 3. Test GetCalendarEvents
	eventsResp, err := client.GetCalendarEventsWithResponse(ctx, "ilife_official", &GetCalendarEventsParams{})
	if err != nil {
		t.Fatalf("GetCalendarEvents failed: %v", err)
	}
	if eventsResp.StatusCode() != http.StatusOK {
		t.Errorf("expected 200, got %d", eventsResp.StatusCode())
	}
	if len(eventsResp.JSON200.Events) != 2 {
		t.Errorf("expected 2 events, got %d", len(eventsResp.JSON200.Events))
	}
	if eventsResp.JSON200.Events[0].Title != "MEGALiFE! 先行物販＠Kアリーナ横浜" {
		t.Errorf("unexpected event title: %s", eventsResp.JSON200.Events[0].Title)
	}
}
