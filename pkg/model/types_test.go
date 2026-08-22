package model

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestNormalizeCalendar(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "sample_calendar.json"))
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var raw RawCalendarResponse
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal raw calendar: %v", err)
	}

	cal := NormalizeCalendar(&raw.PublicCalendar)
	if cal == nil {
		t.Fatal("expected non-nil Calendar")
	}

	if cal.ID != "46438" {
		t.Errorf("expected ID='46438', got '%s'", cal.ID)
	}
	if cal.AliasCode != "ilife_official" {
		t.Errorf("expected AliasCode='ilife_official', got '%s'", cal.AliasCode)
	}
	if cal.Title != "iLiFE!" {
		t.Errorf("expected Title='iLiFE!', got '%s'", cal.Title)
	}
	if cal.Description != "iLiFE!のスケジュールです。" {
		t.Errorf("expected Description, got '%s'", cal.Description)
	}
	if cal.SNSLinks == nil || cal.SNSLinks.Twitter != "https://x.com/iLiFE_official" {
		t.Errorf("expected Twitter SNS link, got '%+v'", cal.SNSLinks)
	}
}

func TestNormalizeEvent(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "testdata", "sample_events.json"))
	if err != nil {
		t.Fatalf("failed to read fixture: %v", err)
	}

	var raw RawEventsResponse
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("failed to unmarshal raw events: %v", err)
	}

	if len(raw.PublicEvents) != 2 {
		t.Fatalf("expected 2 raw events, got %d", len(raw.PublicEvents))
	}

	ev1 := NormalizeEvent(&raw.PublicEvents[0])
	if ev1.ID != "3012407381357599548" {
		t.Errorf("expected ID='3012407381357599548', got '%s'", ev1.ID)
	}
	if ev1.Title != "MEGALiFE! 先行物販＠Kアリーナ横浜" {
		t.Errorf("expected Title, got '%s'", ev1.Title)
	}
	if ev1.Location == nil || *ev1.Location != "Kアリーナ横浜" {
		t.Errorf("expected Location='Kアリーナ横浜', got '%v'", ev1.Location)
	}
	if ev1.URL == nil || *ev1.URL != "https://heroines.jp/news/public/_/8txp75btlwrhdofl.html" {
		t.Errorf("expected URL, got '%v'", ev1.URL)
	}
	if ev1.AllDay {
		t.Error("expected AllDay=false")
	}
	if len(ev1.ImageURLs) != 1 {
		t.Errorf("expected 1 ImageURL, got %d", len(ev1.ImageURLs))
	}

	ev2 := NormalizeEvent(&raw.PublicEvents[1])
	if ev2.ID != "3012407381357599549" {
		t.Errorf("expected ID='3012407381357599549', got '%s'", ev2.ID)
	}
	if !ev2.AllDay {
		t.Error("expected AllDay=true")
	}
}
