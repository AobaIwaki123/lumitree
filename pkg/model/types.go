// Package model defines core domain entities and data normalization logic for lumitree.
package model

import (
	"fmt"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/AobaIwaki123/lumi-spec/event"
)

var eventURLPattern = regexp.MustCompile(`https?://[^\s<>"'。、！？（）「」『』【】]+`)

// Calendar represents normalized calendar metadata.
type Calendar struct {
	ID            string    `json:"id"`
	AliasCode     string    `json:"aliasCode"`
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	CoverImageURL string    `json:"coverImageUrl,omitempty"`
	SNSLinks      *SNSLinks `json:"snsLinks,omitempty"`
}

// SNSLinks holds social media links.
type SNSLinks struct {
	Twitter   string `json:"twitter,omitempty"`
	Instagram string `json:"instagram,omitempty"`
	Website   string `json:"website,omitempty"`
}

// Event represents a normalized single calendar event.
type Event struct {
	ID          string    `json:"id"`
	UUID        string    `json:"uuid,omitempty"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartAt     time.Time `json:"startAt"`
	EndAt       time.Time `json:"endAt"`
	AllDay      bool      `json:"allDay"`
	Timezone    string    `json:"timezone"`
	Location    *string   `json:"location,omitempty"`
	URL         *string   `json:"url,omitempty"`
	ImageURLs   []string  `json:"imageUrls,omitempty"`
	Links       []Link    `json:"links,omitempty"`
}

// Link is an additional URL discovered while normalizing a TimeTree event.
// URL remains the source record's primary URL; Links preserves ticket and
// related destinations found in the event description.
type Link struct {
	Kind   string `json:"kind"`
	URL    string `json:"url"`
	Label  string `json:"label,omitempty"`
	Origin string `json:"origin,omitempty"`
}

// Pagination metadata.
type Pagination struct {
	CurrentPage int `json:"currentPage"`
	TotalPages  int `json:"totalPages"`
	TotalCount  int `json:"totalCount"`
}

// EventListResponse holds normalized calendar and event list.
type EventListResponse struct {
	Calendar   *Calendar   `json:"calendar"`
	Events     []*Event    `json:"events"`
	Pagination *Pagination `json:"pagination"`
}

// RawCalendarResponse represents TimeTree's raw calendar envelope.
type RawCalendarResponse struct {
	PublicCalendar RawCalendar `json:"public_calendar"`
}

// RawCalendar represents TimeTree's raw calendar object.
type RawCalendar struct {
	ID        any                `json:"id"`
	AliasCode string             `json:"alias_code"`
	Name      string             `json:"name"`
	Overview  *string            `json:"overview"`
	Images    *RawImages         `json:"images"`
	Links     map[string]*string `json:"links"`
}

// RawImages represents image container in raw calendar response.
type RawImages struct {
	Cover any `json:"cover"` // Can be *RawImage or []RawImage
}

// RawImage represents a single image URL pair.
type RawImage struct {
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url"`
}

// RawEventsResponse represents TimeTree's raw events envelope.
type RawEventsResponse struct {
	Paging       RawPaging  `json:"paging"`
	PublicEvents []RawEvent `json:"public_events"`
}

// RawPaging represents TimeTree's raw pagination object.
type RawPaging struct {
	CurrentPage int `json:"current_page"`
	TotalPages  int `json:"total_pages"`
	TotalCount  int `json:"total_count"`
}

// RawEvent represents TimeTree's raw event object.
type RawEvent struct {
	ID             any            `json:"id"`
	UUID           string         `json:"uuid"`
	Title          string         `json:"title"`
	Note           *string        `json:"note"`
	Overview       *string        `json:"overview"`
	Description    *string        `json:"description"`
	StartAt        int64          `json:"start_at"`
	EndAt          int64          `json:"end_at"`
	AllDay         bool           `json:"all_day"`
	StartTimestamp *int64         `json:"start_timestamp"`
	StartTimezone  *string        `json:"start_timezone"`
	EndTimezone    *string        `json:"end_timezone"`
	RegionTimezone *string        `json:"region_timezone"`
	LocationName   *string        `json:"location_name"`
	Location       *string        `json:"location"`
	LinkURL        *string        `json:"link_url"`
	URL            *string        `json:"url"`
	Images         map[string]any `json:"images"`
}

func parseID(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return fmt.Sprintf("%.0f", val)
	case int:
		return strconv.Itoa(val)
	case int64:
		return strconv.FormatInt(val, 10)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// NormalizeCalendar converts a TimeTree RawCalendar to a clean model.Calendar.
func NormalizeCalendar(raw *RawCalendar) *Calendar {
	if raw == nil {
		return nil
	}

	title := strings.TrimSpace(raw.Name)
	if title == "" {
		title = raw.AliasCode
	}

	desc := ""
	if raw.Overview != nil {
		desc = strings.TrimSpace(*raw.Overview)
	}

	var coverURL string
	if raw.Images != nil && raw.Images.Cover != nil {
		switch c := raw.Images.Cover.(type) {
		case map[string]any:
			if u, ok := c["url"].(string); ok {
				coverURL = u
			}
		case []any:
			if len(c) > 0 {
				if m, ok := c[0].(map[string]any); ok {
					if u, ok := m["url"].(string); ok {
						coverURL = u
					}
				}
			}
		}
	}

	sns := &SNSLinks{}
	if raw.Links != nil {
		if twitter, ok := raw.Links["twitter"]; ok && twitter != nil {
			sns.Twitter = *twitter
		}
		if instagram, ok := raw.Links["instagram"]; ok && instagram != nil {
			sns.Instagram = *instagram
		}
		if website, ok := raw.Links["website"]; ok && website != nil {
			sns.Website = *website
		}
	}

	return &Calendar{
		ID:            parseID(raw.ID),
		AliasCode:     raw.AliasCode,
		Title:         title,
		Description:   desc,
		CoverImageURL: coverURL,
		SNSLinks:      sns,
	}
}

// NormalizeEvent converts a TimeTree RawEvent to a clean model.Event.
func NormalizeEvent(raw *RawEvent) *Event {
	if raw == nil {
		return nil
	}

	tzStr := "Asia/Tokyo"
	if raw.RegionTimezone != nil && *raw.RegionTimezone != "" {
		tzStr = *raw.RegionTimezone
	} else if raw.StartTimezone != nil && *raw.StartTimezone != "" && *raw.StartTimezone != "UTC" {
		tzStr = *raw.StartTimezone
	}

	loc, err := time.LoadLocation(tzStr)
	if err != nil {
		loc = time.FixedZone("JST", 9*60*60)
	}

	// TimeTree timestamps are in milliseconds
	startAt := time.UnixMilli(raw.StartAt).In(loc)
	endAt := time.UnixMilli(raw.EndAt).In(loc)

	// If endAt is equal to or before startAt on a non-allday event, default to startAt + 1 hour
	if !raw.AllDay && (endAt.Before(startAt) || endAt.Equal(startAt)) {
		endAt = startAt.Add(1 * time.Hour)
	}

	// Description resolution: note > overview > description
	desc := ""
	if raw.Note != nil && strings.TrimSpace(*raw.Note) != "" {
		desc = strings.TrimSpace(*raw.Note)
	} else if raw.Overview != nil && strings.TrimSpace(*raw.Overview) != "" {
		desc = strings.TrimSpace(*raw.Overview)
	} else if raw.Description != nil && strings.TrimSpace(*raw.Description) != "" {
		desc = strings.TrimSpace(*raw.Description)
	}

	// Location resolution: location_name > location
	var locPtr *string
	if raw.LocationName != nil && strings.TrimSpace(*raw.LocationName) != "" {
		trimmed := strings.TrimSpace(*raw.LocationName)
		locPtr = &trimmed
	} else if raw.Location != nil && strings.TrimSpace(*raw.Location) != "" {
		trimmed := strings.TrimSpace(*raw.Location)
		locPtr = &trimmed
	}

	// URL resolution: link_url > url
	var urlPtr *string
	if raw.LinkURL != nil && strings.TrimSpace(*raw.LinkURL) != "" {
		trimmed := strings.TrimSpace(*raw.LinkURL)
		urlPtr = &trimmed
	} else if raw.URL != nil && strings.TrimSpace(*raw.URL) != "" {
		trimmed := strings.TrimSpace(*raw.URL)
		urlPtr = &trimmed
	}

	// Images extraction
	var imageURLs []string
	if raw.Images != nil {
		for _, imgGroup := range raw.Images {
			if list, ok := imgGroup.([]any); ok {
				for _, item := range list {
					if m, ok := item.(map[string]any); ok {
						if u, ok := m["url"].(string); ok && u != "" {
							imageURLs = append(imageURLs, u)
						}
					}
				}
			}
		}
	}

	links := extractLinks(desc, urlPtr)

	return &Event{
		ID:          parseID(raw.ID),
		UUID:        raw.UUID,
		Title:       strings.TrimSpace(raw.Title),
		Description: desc,
		StartAt:     startAt,
		EndAt:       endAt,
		AllDay:      raw.AllDay,
		Timezone:    tzStr,
		Location:    locPtr,
		URL:         urlPtr,
		ImageURLs:   imageURLs,
		Links:       links,
	}
}

func extractLinks(description string, primaryURL *string) []Link {
	links := make([]Link, 0, 1)
	seen := make(map[string]struct{})
	add := func(rawURL, label, origin string) {
		rawURL = cleanURL(rawURL)
		if rawURL == "" {
			return
		}
		parsed, err := url.Parse(rawURL)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" {
			return
		}
		key := strings.ToLower(rawURL)
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		kind := classifyLink(rawURL, label)
		if origin == "structured_url" && kind == "related" {
			kind = "official"
		}
		links = append(links, Link{Kind: kind, URL: rawURL, Label: strings.TrimSpace(label), Origin: origin})
	}
	if primaryURL != nil {
		add(*primaryURL, "", "structured_url")
	}
	for _, rawURL := range eventURLPattern.FindAllString(description, -1) {
		add(rawURL, descriptionContext(description, rawURL), "description")
	}
	return links
}

func cleanURL(rawURL string) string {
	return strings.TrimRight(strings.TrimSpace(rawURL), ".,!?;:。、)]}〉》」』】")
}

func descriptionContext(description, rawURL string) string {
	index := strings.Index(description, rawURL)
	if index < 0 {
		return ""
	}
	start := index - 20
	if start < 0 {
		start = 0
	}
	return strings.TrimSpace(description[start:index])
}

func classifyLink(rawURL, label string) string {
	parsed, _ := url.Parse(rawURL)
	host := strings.ToLower(strings.TrimPrefix(parsed.Hostname(), "www."))
	text := strings.ToLower(label + " " + rawURL)
	if strings.Contains(text, "ticket") || strings.Contains(text, "チケット") || strings.Contains(text, "受付") || strings.Contains(text, "販売") || strings.Contains(text, "先行") || strings.Contains(text, "一般発売") {
		return "ticket"
	}
	switch host {
	case "ticketdive.com", "t-dv.com", "tiget.net", "eplus.jp", "l-tike.com", "livepocket.jp", "passmarket.yahoo.co.jp", "ticketpay.jp", "teket.jp":
		return "ticket"
	case "x.com", "twitter.com", "instagram.com", "tiktok.com", "youtube.com", "youtu.be", "twitch.tv":
		return "social"
	}
	if strings.Contains(text, "公式") || strings.Contains(text, "official") || strings.Contains(text, "ホームページ") || strings.Contains(text, "website") {
		return "official"
	}
	return "related"
}

// ToEventCandidate maps one normalized TimeTree event to lumi-spec.
func ToEventCandidate(calendarID string, item *Event) event.Candidate {
	if item == nil {
		return event.Candidate{}
	}
	localDate := item.StartAt.Format("2006-01-02")
	recordURL := ""
	if item.URL != nil {
		recordURL = *item.URL
	}
	links := make([]event.Link, 0, len(item.Links))
	for _, link := range item.Links {
		links = append(links, event.Link{Kind: link.Kind, URL: link.URL, Label: link.Label, Origin: link.Origin})
	}
	return event.Candidate{
		SchemaVersion: event.CandidateSchemaVersion,
		ID:            "lumitree:" + item.ID,
		SourceRef: event.SourceRef{
			Source:       "lumitree",
			RecordID:     item.ID,
			URL:          recordURL,
			CollectionID: calendarID,
		},
		Title:     item.Title,
		LocalDate: localDate,
		StartsAt:  item.StartAt.Format(time.RFC3339),
		EndsAt:    item.EndAt.Format(time.RFC3339),
		Timezone:  item.Timezone,
		Venue:     stringValue(item.Location),
		Links:     links,
	}
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
