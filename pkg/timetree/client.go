// Package timetree implements an unofficial client for TimeTree public calendar web API.
package timetree

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AobaIwaki123/lumitree/pkg/model"
)

const (
	DefaultUserAgent   = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
	DefaultTimeTreeApp = "web/2.1.0/1.0.0"
)

var csrfMetaRegex = regexp.MustCompile(`<meta\s+name=["']csrf-token["']\s+content=["']([^"']+)["']`)

// Client is a TimeTree public calendar API client with automated CSRF handshake and session handling.
type Client struct {
	baseURL    string
	httpClient *http.Client
	logger     *slog.Logger
	userAgent  string
	appHeader  string

	mu         sync.RWMutex
	csrfTokens map[string]string // calendarID -> csrfToken
}

// Option configures Client.
type Option func(*Client)

// WithBaseURL sets a custom base URL (useful for testing).
func WithBaseURL(u string) Option {
	return func(c *Client) {
		c.baseURL = strings.TrimRight(u, "/")
	}
}

// WithHTTPClient sets a custom http.Client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// WithLogger sets a custom logger.
func WithLogger(l *slog.Logger) Option {
	return func(c *Client) {
		c.logger = l
	}
}

// NewClient creates a new TimeTree client.
func NewClient(opts ...Option) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookiejar: %w", err)
	}

	c := &Client{
		baseURL: "https://timetreeapp.com",
		httpClient: &http.Client{
			Jar:     jar,
			Timeout: 15 * time.Second,
		},
		logger:     slog.Default(),
		userAgent:  DefaultUserAgent,
		appHeader:  DefaultTimeTreeApp,
		csrfTokens: make(map[string]string),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// ensureSession fetches the public calendar web page to capture session cookies and the CSRF token.
func (c *Client) ensureSession(ctx context.Context, calendarID string) (string, error) {
	c.mu.RLock()
	token, exists := c.csrfTokens[calendarID]
	c.mu.RUnlock()
	if exists && token != "" {
		return token, nil
	}

	pageURL := fmt.Sprintf("%s/public_calendars/%s", c.baseURL, calendarID)
	c.logger.DebugContext(ctx, "fetching initial calendar HTML for session handshake", "url", pageURL)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request for initial page: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch initial calendar page: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("initial page returned HTTP %d %s", resp.StatusCode, resp.Status)
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read initial page body: %w", err)
	}

	matches := csrfMetaRegex.FindSubmatch(bodyBytes)
	if len(matches) < 2 {
		return "", fmt.Errorf("csrf-token meta tag not found in page HTML")
	}

	extractedToken := string(matches[1])
	c.mu.Lock()
	c.csrfTokens[calendarID] = extractedToken
	c.mu.Unlock()

	c.logger.DebugContext(ctx, "successfully acquired CSRF token", "calendar_id", calendarID)
	return extractedToken, nil
}

// GetCalendar fetches calendar metadata and returns a normalized model.Calendar.
func (c *Client) GetCalendar(ctx context.Context, calendarID string) (*model.Calendar, error) {
	csrfToken, err := c.ensureSession(ctx, calendarID)
	if err != nil {
		return nil, fmt.Errorf("session preparation failed: %w", err)
	}

	apiURL := fmt.Sprintf("%s/api/v2/public_calendars/%s", c.baseURL, calendarID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create calendar API request: %w", err)
	}

	c.setHeaders(req, calendarID, csrfToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("calendar API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("calendar API returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var raw model.RawCalendarResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode calendar API response: %w", err)
	}

	return model.NormalizeCalendar(&raw.PublicCalendar), nil
}

// GetEvents fetches event list and returns normalized events.
func (c *Client) GetEvents(ctx context.Context, calendarID string, year, month, page int) (*model.EventListResponse, error) {
	csrfToken, err := c.ensureSession(ctx, calendarID)
	if err != nil {
		return nil, fmt.Errorf("session preparation failed: %w", err)
	}

	apiURL := fmt.Sprintf("%s/api/v2/public_calendars/%s/public_events", c.baseURL, calendarID)
	q := url.Values{}
	if year > 0 {
		q.Set("year", strconv.Itoa(year))
	}
	if month > 0 {
		q.Set("month", strconv.Itoa(month))
	}
	if page > 0 {
		q.Set("page", strconv.Itoa(page))
	}
	if len(q) > 0 {
		apiURL += "?" + q.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create events API request: %w", err)
	}

	c.setHeaders(req, calendarID, csrfToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("events API request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("events API returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	var raw model.RawEventsResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("failed to decode events API response: %w", err)
	}

	events := make([]*model.Event, 0, len(raw.PublicEvents))
	for i := range raw.PublicEvents {
		events = append(events, model.NormalizeEvent(&raw.PublicEvents[i]))
	}

	// Fetch calendar info to bundle into EventListResponse
	cal, err := c.GetCalendar(ctx, calendarID)
	if err != nil {
		// Fallback minimal calendar info if metadata fetch fails
		cal = &model.Calendar{
			ID:        calendarID,
			AliasCode: calendarID,
			Title:     calendarID,
		}
	}

	return &model.EventListResponse{
		Calendar: cal,
		Events:   events,
		Pagination: &model.Pagination{
			CurrentPage: raw.Paging.CurrentPage,
			TotalPages:  raw.Paging.TotalPages,
			TotalCount:  raw.Paging.TotalCount,
		},
	}, nil
}

func (c *Client) setHeaders(req *http.Request, calendarID, csrfToken string) {
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Referer", fmt.Sprintf("%s/public_calendars/%s", c.baseURL, calendarID))
	req.Header.Set("X-CSRF-Token", csrfToken)
	req.Header.Set("X-TimeTreeA", c.appHeader)
}
