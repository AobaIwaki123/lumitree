package timetree

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestClientGetCalendarAndEvents(t *testing.T) {
	pageHTML, err := os.ReadFile(filepath.Join("..", "..", "testdata", "sample_page.html"))
	if err != nil {
		t.Fatalf("failed to read sample_page.html: %v", err)
	}

	calJSON, err := os.ReadFile(filepath.Join("..", "..", "testdata", "sample_calendar.json"))
	if err != nil {
		t.Fatalf("failed to read sample_calendar.json: %v", err)
	}

	eventsJSON, err := os.ReadFile(filepath.Join("..", "..", "testdata", "sample_events.json"))
	if err != nil {
		t.Fatalf("failed to read sample_events.json: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/public_calendars/ilife_official":
			http.SetCookie(w, &http.Cookie{
				Name:  "_session_id",
				Value: "mock_session_12345",
				Path:  "/",
			})
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(pageHTML)

		case "/api/v2/public_calendars/ilife_official":
			// Verify required headers
			if r.Header.Get("X-CSRF-Token") != "mock_csrf_token_abcdef1234567890" {
				t.Errorf("expected X-CSRF-Token header, got '%s'", r.Header.Get("X-CSRF-Token"))
				http.Error(w, "invalid csrf token", http.StatusBadRequest)
				return
			}
			if r.Header.Get("X-TimeTreeA") != "web/2.1.0/1.0.0" {
				t.Errorf("expected X-TimeTreeA header, got '%s'", r.Header.Get("X-TimeTreeA"))
			}

			// Verify cookie was sent
			cookie, err := r.Cookie("_session_id")
			if err != nil || cookie.Value != "mock_session_12345" {
				t.Errorf("expected _session_id cookie, got: %v", cookie)
			}

			w.Header().Set("Content-Type", "application/json")
			w.Write(calJSON)

		case "/api/v2/public_calendars/ilife_official/public_events":
			w.Header().Set("Content-Type", "application/json")
			w.Write(eventsJSON)

		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client, err := NewClient(WithBaseURL(server.URL))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	ctx := context.Background()

	// Test GetCalendar
	cal, err := client.GetCalendar(ctx, "ilife_official")
	if err != nil {
		t.Fatalf("GetCalendar failed: %v", err)
	}
	if cal.Title != "iLiFE!" {
		t.Errorf("expected Title='iLiFE!', got '%s'", cal.Title)
	}
	if cal.AliasCode != "ilife_official" {
		t.Errorf("expected AliasCode='ilife_official', got '%s'", cal.AliasCode)
	}

	// Test GetEvents
	eventList, err := client.GetEvents(ctx, "ilife_official", 2026, 8, 1)
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}
	if len(eventList.Events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(eventList.Events))
	}
	if eventList.Events[0].Title != "MEGALiFE! 先行物販＠Kアリーナ横浜" {
		t.Errorf("unexpected title: %s", eventList.Events[0].Title)
	}
	if eventList.Events[0].ID != "3012407381357599548" {
		t.Errorf("unexpected ID: %s", eventList.Events[0].ID)
	}
}
