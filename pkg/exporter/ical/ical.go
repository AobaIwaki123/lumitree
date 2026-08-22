// Package ical provides RFC 5545 compliant iCalendar generation for lumitree.
package ical

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/AobaIwaki123/lumitree/pkg/model"
)

// Generate builds an RFC 5545 compliant iCalendar byte stream from calendar metadata and events.
func Generate(cal *model.Calendar, events []*model.Event) ([]byte, error) {
	var buf bytes.Buffer

	calTitle := "lumitree calendar"
	calID := "default"
	if cal != nil {
		if cal.Title != "" {
			calTitle = cal.Title
		}
		if cal.AliasCode != "" {
			calID = cal.AliasCode
		} else if cal.ID != "" {
			calID = cal.ID
		}
	}

	buf.WriteString("BEGIN:VCALENDAR\r\n")
	buf.WriteString("VERSION:2.0\r\n")
	buf.WriteString("PRODID:-//lumitree//lumitree 1.0.0//EN\r\n")
	buf.WriteString("CALSCALE:GREGORIAN\r\n")
	buf.WriteString("METHOD:PUBLISH\r\n")
	buf.WriteString(fmt.Sprintf("X-WR-CALNAME:%s\r\n", escapeText(calTitle)))
	buf.WriteString("X-WR-TIMEZONE:Asia/Tokyo\r\n")

	nowUTC := time.Now().UTC().Format("20060102T150405Z")

	for _, ev := range events {
		if ev == nil {
			continue
		}

		buf.WriteString("BEGIN:VEVENT\r\n")
		uid := fmt.Sprintf("%s-%s@lumitree", calID, ev.ID)
		if ev.UUID != "" {
			uid = fmt.Sprintf("%s-%s@lumitree", calID, ev.UUID)
		}
		buf.WriteString(fmt.Sprintf("UID:%s\r\n", uid))
		buf.WriteString(fmt.Sprintf("DTSTAMP:%s\r\n", nowUTC))
		buf.WriteString(fmt.Sprintf("SUMMARY:%s\r\n", escapeText(ev.Title)))

		if ev.Description != "" {
			buf.WriteString(fmt.Sprintf("DESCRIPTION:%s\r\n", escapeText(ev.Description)))
		}

		if ev.Location != nil && *ev.Location != "" {
			buf.WriteString(fmt.Sprintf("LOCATION:%s\r\n", escapeText(*ev.Location)))
		}

		if ev.URL != nil && *ev.URL != "" {
			buf.WriteString(fmt.Sprintf("URL:%s\r\n", *ev.URL))
		}

		if ev.AllDay {
			// All day event: VALUE=DATE:YYYYMMDD, DTEND is next day
			dtStart := ev.StartAt.Format("20060102")
			dtEnd := ev.EndAt.AddDate(0, 0, 1).Format("20060102")
			buf.WriteString(fmt.Sprintf("DTSTART;VALUE=DATE:%s\r\n", dtStart))
			buf.WriteString(fmt.Sprintf("DTEND;VALUE=DATE:%s\r\n", dtEnd))
		} else {
			// Timed event: TZID=Asia/Tokyo:YYYYMMDDTHHMMSS
			tz := "Asia/Tokyo"
			if ev.Timezone != "" {
				tz = ev.Timezone
			}
			dtStart := ev.StartAt.Format("20060102T150405")
			dtEnd := ev.EndAt.Format("20060102T150405")
			buf.WriteString(fmt.Sprintf("DTSTART;TZID=%s:%s\r\n", tz, dtStart))
			buf.WriteString(fmt.Sprintf("DTEND;TZID=%s:%s\r\n", tz, dtEnd))
		}

		buf.WriteString("END:VEVENT\r\n")
	}

	buf.WriteString("END:VCALENDAR\r\n")

	return buf.Bytes(), nil
}

func escapeText(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, ";", "\\;")
	s = strings.ReplaceAll(s, ",", "\\,")
	s = strings.ReplaceAll(s, "\r\n", "\\n")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\n")
	return s
}
