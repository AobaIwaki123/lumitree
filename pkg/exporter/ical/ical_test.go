package ical

import (
	"strings"
	"testing"
	"time"

	"github.com/AobaIwaki123/lumitree/pkg/model"
)

func TestGenerateICal(t *testing.T) {
	loc := time.FixedZone("JST", 9*60*60)
	locStr := "Kアリーナ横浜"
	urlStr := "https://ticket.heroines.jp/megalife"

	cal := &model.Calendar{
		ID:        "46438",
		AliasCode: "ilife_official",
		Title:     "iLiFE!",
	}

	events := []*model.Event{
		{
			ID:          "10001",
			UUID:        "uuid-1",
			Title:       "MEGALiFE! 先行物販＠Kアリーナ横浜",
			Description: "物販のご案内",
			StartAt:     time.Date(2026, 8, 25, 9, 0, 0, 0, loc),
			EndAt:       time.Date(2026, 8, 25, 12, 0, 0, 0, loc),
			AllDay:      false,
			Timezone:    "Asia/Tokyo",
			Location:    &locStr,
			URL:         &urlStr,
		},
		{
			ID:          "10002",
			UUID:        "uuid-2",
			Title:       "若葉のあ生誕祭",
			Description: "生誕祭イベント",
			StartAt:     time.Date(2026, 9, 27, 0, 0, 0, 0, loc),
			EndAt:       time.Date(2026, 9, 27, 0, 0, 0, 0, loc),
			AllDay:      true,
			Timezone:    "Asia/Tokyo",
		},
	}

	data, err := Generate(cal, events)
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	output := string(data)

	if !strings.Contains(output, "BEGIN:VCALENDAR") || !strings.Contains(output, "END:VCALENDAR") {
		t.Error("expected VCALENDAR envelope")
	}
	if !strings.Contains(output, "X-WR-CALNAME:iLiFE!") {
		t.Errorf("expected calendar name, got: %s", output)
	}
	if !strings.Contains(output, "SUMMARY:MEGALiFE! 先行物販＠Kアリーナ横浜") {
		t.Errorf("expected event 1 summary, got: %s", output)
	}
	if !strings.Contains(output, "DTSTART;TZID=Asia/Tokyo:20260825T090000") {
		t.Errorf("expected DTSTART for timed event, got: %s", output)
	}
	if !strings.Contains(output, "LOCATION:Kアリーナ横浜") {
		t.Errorf("expected LOCATION, got: %s", output)
	}
	if !strings.Contains(output, "DTSTART;VALUE=DATE:20260927") {
		t.Errorf("expected DTSTART for all-day event, got: %s", output)
	}
	if !strings.Contains(output, "DTEND;VALUE=DATE:20260928") {
		t.Errorf("expected DTEND for all-day event (+1 day), got: %s", output)
	}
}
