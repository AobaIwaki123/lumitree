// Package main is the entrypoint for lumitree CLI and server.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/AobaIwaki123/lumitree/pkg/config"
	"github.com/AobaIwaki123/lumitree/pkg/exporter/ical"
	"github.com/AobaIwaki123/lumitree/pkg/logger"
	"github.com/AobaIwaki123/lumitree/pkg/server"
	"github.com/AobaIwaki123/lumitree/pkg/timetree"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	cfg := config.Load()
	_ = logger.Init(cfg.LogLevel, cfg.LogFormat)

	switch command {
	case "serve":
		runServe(cfg, args)
	case "get":
		runGet(cfg, args)
	case "ics":
		runICS(cfg, args)
	case "version":
		fmt.Printf("lumitree version %s (commit: %s, built at: %s)\n", version, commit, date)
	case "help", "-h", "--help":
		printUsage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: lumitree <command> [arguments]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  serve               Start the HTTP webcal / OpenAPI proxy server")
	fmt.Println("  get <calendar-id>   Fetch and display events for a TimeTree calendar")
	fmt.Println("  ics <calendar-id>   Export calendar events to an RFC 5545 iCalendar (.ics) file")
	fmt.Println("  version             Print version and build information")
	fmt.Println("  help                Show this help message")
}

func runServe(cfg *config.Config, args []string) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--port", "-p":
			if i+1 < len(args) {
				p, err := strconv.Atoi(args[i+1])
				if err == nil {
					cfg.Port = p
				}
				i++
			}
		case "--host":
			if i+1 < len(args) {
				cfg.Host = args[i+1]
				i++
			}
		}
	}

	client, err := timetree.NewClient(timetree.WithBaseURL(cfg.TimeTreeBaseURL))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create timetree client: %v\n", err)
		os.Exit(1)
	}

	srv := server.NewServer(cfg, client, nil)
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	fmt.Printf("Starting lumitree HTTP proxy server on %s ...\n", addr)

	httpServer := &http.Server{
		Addr:              addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		fmt.Fprintf(os.Stderr, "server failed: %v\n", err)
		os.Exit(1)
	}
}

func runGet(cfg *config.Config, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: calendar-id is required")
		fmt.Fprintln(os.Stderr, "Usage: lumitree get <calendar-id> [--json] [--year YYYY]")
		os.Exit(1)
	}

	calendarID := args[0]
	isJSON := false
	year := time.Now().Year()

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--json":
			isJSON = true
		case "--year":
			if i+1 < len(args) {
				y, err := strconv.Atoi(args[i+1])
				if err == nil {
					year = y
				}
				i++
			}
		}
	}

	client, err := timetree.NewClient(timetree.WithBaseURL(cfg.TimeTreeBaseURL))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create timetree client: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	eventList, err := client.GetEvents(ctx, calendarID, year, 0, 1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get events: %v\n", err)
		os.Exit(1)
	}

	if isJSON {
		data, _ := json.MarshalIndent(eventList, "", "  ")
		fmt.Println(string(data))
		return
	}

	cal := eventList.Calendar
	events := eventList.Events
	fmt.Printf("Calendar: %s (%s)\n", cal.Title, cal.ID)
	if cal.Description != "" {
		fmt.Printf("Description: %s\n", cal.Description)
	}
	fmt.Printf("Total Events (%d): %d\n\n", year, len(events))

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "START\tEND\tTITLE\tLOCATION")
	for _, ev := range events {
		startStr := ev.StartAt.Format("2006-01-02 15:04")
		if ev.AllDay {
			startStr = ev.StartAt.Format("2006-01-02 (All Day)")
		}
		endStr := ev.EndAt.Format("15:04")
		if ev.AllDay {
			endStr = "-"
		}
		locStr := ""
		if ev.Location != nil {
			locStr = *ev.Location
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", startStr, endStr, ev.Title, locStr)
	}
	_ = w.Flush()
}

func runICS(cfg *config.Config, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Error: calendar-id is required")
		fmt.Fprintln(os.Stderr, "Usage: lumitree ics <calendar-id> [--output <file.ics>] [--year YYYY]")
		os.Exit(1)
	}

	calendarID := args[0]
	outputPath := fmt.Sprintf("%s.ics", calendarID)
	year := time.Now().Year()

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--output", "-o":
			if i+1 < len(args) {
				outputPath = args[i+1]
				i++
			}
		case "--year":
			if i+1 < len(args) {
				y, err := strconv.Atoi(args[i+1])
				if err == nil {
					year = y
				}
				i++
			}
		}
	}

	client, err := timetree.NewClient(timetree.WithBaseURL(cfg.TimeTreeBaseURL))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create timetree client: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()
	eventList, err := client.GetEvents(ctx, calendarID, year, 0, 1)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get events: %v\n", err)
		os.Exit(1)
	}

	icsData, err := ical.Generate(eventList.Calendar, eventList.Events)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate iCal: %v\n", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil && filepath.Dir(outputPath) != "." {
		fmt.Fprintf(os.Stderr, "failed to create directory: %v\n", err)
		os.Exit(1)
	}

	if err := os.WriteFile(outputPath, icsData, 0o600); err != nil {
		fmt.Fprintf(os.Stderr, "failed to write ics file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully exported %d events to %s\n", len(eventList.Events), outputPath)
}
