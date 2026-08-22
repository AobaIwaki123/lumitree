package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestLoggerJSON(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New("info", "json", buf)

	l.InfoContext(context.Background(), "test message", "calendar_id", "ilife_official", "count", 12)

	var data map[string]interface{}
	if err := json.Unmarshal(buf.Bytes(), &data); err != nil {
		t.Fatalf("failed to unmarshal log json: %v, raw: %s", err, buf.String())
	}

	if data["msg"] != "test message" {
		t.Errorf("expected msg='test message', got '%v'", data["msg"])
	}
	if data["calendar_id"] != "ilife_official" {
		t.Errorf("expected calendar_id='ilife_official', got '%v'", data["calendar_id"])
	}
	if data["count"] != float64(12) {
		t.Errorf("expected count=12, got '%v'", data["count"])
	}
}

func TestLoggerText(t *testing.T) {
	buf := &bytes.Buffer{}
	l := New("debug", "text", buf)

	l.DebugContext(context.Background(), "debug event", "key", "val")

	output := buf.String()
	if !strings.Contains(output, "level=DEBUG") {
		t.Errorf("expected level=DEBUG in output, got: %s", output)
	}
	if !strings.Contains(output, "debug event") {
		t.Errorf("expected 'debug event' in output, got: %s", output)
	}
}
