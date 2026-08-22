// Package server implements OpenAPI compliant HTTP proxy server for lumitree.
package server

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AobaIwaki123/lumitree/pkg/cache"
	"github.com/AobaIwaki123/lumitree/pkg/config"
	"github.com/AobaIwaki123/lumitree/pkg/exporter/ical"
	"github.com/AobaIwaki123/lumitree/pkg/model"
	"github.com/AobaIwaki123/lumitree/pkg/timetree"
)

// Server is the HTTP server for lumitree proxy.
type Server struct {
	cfg    *config.Config
	client *timetree.Client
	cache  *cache.MemoryCache
	logger *slog.Logger
	mux    *http.ServeMux
}

// NewServer creates a new configured HTTP Server.
func NewServer(cfg *config.Config, client *timetree.Client, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}

	s := &Server{
		cfg:    cfg,
		client: client,
		cache:  cache.NewMemoryCache(cfg.CacheTTL),
		logger: logger,
		mux:    http.NewServeMux(),
	}

	s.routes()
	return s
}

// Handler returns the HTTP handler with middleware.
func (s *Server) Handler() http.Handler {
	return s.loggingMiddleware(s.mux)
}

// routes registers all API routes.
func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealthz)
	s.mux.HandleFunc("GET /api/v1/calendars/{calendarId}", s.handleGetCalendar)
	s.mux.HandleFunc("GET /api/v1/calendars/{calendarId}/events", s.handleGetEvents)
	s.mux.HandleFunc("GET /api/v1/calendars/{calendarId}/events.ics", s.handleGetEventsICS)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	resp := map[string]any{
		"status":    "ok",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	writeJSON(w, http.StatusOK, resp)
}

func (s *Server) handleGetCalendar(w http.ResponseWriter, r *http.Request) {
	calendarID := r.PathValue("calendarId")
	if calendarID == "" {
		s.writeError(w, r, http.StatusBadRequest, "Invalid Calendar ID", "calendarId path parameter is required")
		return
	}

	cacheKey := fmt.Sprintf("cal:%s", calendarID)
	if cached, ok := s.cache.Get(cacheKey); ok {
		if cal, ok := cached.(*model.Calendar); ok {
			writeJSON(w, http.StatusOK, cal)
			return
		}
	}

	cal, err := s.client.GetCalendar(r.Context(), calendarID)
	if err != nil {
		s.logger.ErrorContext(r.Context(), "failed to fetch calendar", "calendar_id", calendarID, "error", err)
		s.writeError(w, r, http.StatusBadGateway, "Upstream Error", fmt.Sprintf("Failed to fetch calendar from TimeTree: %v", err))
		return
	}

	s.cache.Set(cacheKey, cal)
	writeJSON(w, http.StatusOK, cal)
}

func (s *Server) handleGetEvents(w http.ResponseWriter, r *http.Request) {
	calendarID := r.PathValue("calendarId")
	if calendarID == "" {
		s.writeError(w, r, http.StatusBadRequest, "Invalid Calendar ID", "calendarId path parameter is required")
		return
	}

	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page <= 0 {
		page = 1
	}

	cacheKey := fmt.Sprintf("events:%s:%d:%d:%d", calendarID, year, month, page)
	if cached, ok := s.cache.Get(cacheKey); ok {
		if eventList, ok := cached.(*model.EventListResponse); ok {
			writeJSON(w, http.StatusOK, eventList)
			return
		}
	}

	eventList, err := s.client.GetEvents(r.Context(), calendarID, year, month, page)
	if err != nil {
		s.logger.ErrorContext(r.Context(), "failed to fetch events", "calendar_id", calendarID, "error", err)
		s.writeError(w, r, http.StatusBadGateway, "Upstream Error", fmt.Sprintf("Failed to fetch events from TimeTree: %v", err))
		return
	}

	s.cache.Set(cacheKey, eventList)
	writeJSON(w, http.StatusOK, eventList)
}

func (s *Server) handleGetEventsICS(w http.ResponseWriter, r *http.Request) {
	calendarID := r.PathValue("calendarId")
	if calendarID == "" {
		http.Error(w, "calendarId required", http.StatusBadRequest)
		return
	}

	cacheKey := fmt.Sprintf("ics:%s", calendarID)
	if cached, ok := s.cache.Get(cacheKey); ok {
		if icsBytes, ok := cached.([]byte); ok {
			w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
			w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s.ics\"", calendarID))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(icsBytes)
			return
		}
	}

	eventList, err := s.client.GetEvents(r.Context(), calendarID, 0, 0, 1)
	if err != nil {
		s.logger.ErrorContext(r.Context(), "failed to fetch events for ICS", "calendar_id", calendarID, "error", err)
		http.Error(w, fmt.Sprintf("Upstream Error: %v", err), http.StatusBadGateway)
		return
	}

	icsBytes, err := ical.Generate(eventList.Calendar, eventList.Events)
	if err != nil {
		s.logger.ErrorContext(r.Context(), "failed to generate iCalendar", "error", err)
		http.Error(w, "Failed to generate iCalendar", http.StatusInternalServerError)
		return
	}

	s.cache.Set(cacheKey, icsBytes)
	w.Header().Set("Content-Type", "text/calendar; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s.ics\"", calendarID))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(icsBytes)
}

func (s *Server) writeError(w http.ResponseWriter, r *http.Request, status int, title, detail string) {
	resp := map[string]any{
		"type":     "https://lumitree.dev/errors/" + strings.ToLower(strings.ReplaceAll(title, " ", "-")),
		"title":    title,
		"status":   status,
		"detail":   detail,
		"instance": r.URL.Path,
	}
	writeJSON(w, status, resp)
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	_ = enc.Encode(data)
}

type responseWriterWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriterWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (s *Server) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriterWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		latency := time.Since(start)
		s.logger.InfoContext(r.Context(), "http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.statusCode,
			"latency_ms", latency.Milliseconds(),
			"remote_addr", r.RemoteAddr,
		)
	})
}
